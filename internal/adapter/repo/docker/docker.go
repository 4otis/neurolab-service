package docker

import (
	"bytes"
	"context"
	"fmt"

	"github.com/4otis/neurolab-service/internal/port/repo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var _ repo.DockerRepo = (*DockerRepo)(nil)

type DockerRepo struct {
	dockerClient *client.Client
}

func NewDockerRepo(dockerClient *client.Client) *DockerRepo {
	return &DockerRepo{dockerClient: dockerClient}
}

func (r *DockerRepo) CreateContainer(ctx context.Context, imageName string, mnts []*string) (respID string, err error) {
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   []string{"sleep", "infinity"},
	}

	m := make([]mount.Mount, 0, len(mnts))
	for i, path := range mnts {
		var target string
		if i == 0 { // first element should be solution
			target = "/app/solutions"
		} else {
			target = "/app/scripts"
		}
		m = append(m, mount.Mount{
			Type:     mount.TypeBind,
			Source:   *path,
			Target:   target,
			ReadOnly: true,
		})
	}

	hostConfig := &container.HostConfig{
		AutoRemove: false,
		Mounts:     m,
	}

	resp, err := r.dockerClient.ContainerCreate(ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		"",
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container using docker client: %w", err)
	}

	if err := r.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		r.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return "", fmt.Errorf("container start failed: %w", err)
	}

	return resp.ID, nil
}

func (r *DockerRepo) DeleteContainer(ctx context.Context, containerID string) error {
	r.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})

	return nil
}

func (r *DockerRepo) ExecCommand(ctx context.Context, containerID string, cmd []string) (repo.ExecCommandResp, error) {
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execResp, err := r.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return repo.ExecCommandResp{
			StatusCode: -1,
		}, fmt.Errorf("failed to create container exec: %w", err)
	}

	attachResp, err := r.dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return repo.ExecCommandResp{
			StatusCode: -1,
		}, fmt.Errorf("failed to attach container exec: %w", err)
	}
	defer attachResp.Close()

	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)
	go func() {
		defer close(outputDone)
		_, err := stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return repo.ExecCommandResp{StatusCode: -1},
				fmt.Errorf("failed to copy multiplexed stream: %w", err)
		}
	case <-ctx.Done():
		return repo.ExecCommandResp{StatusCode: -1}, ctx.Err()
	}

	inspectResp, err := r.dockerClient.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return repo.ExecCommandResp{
			StdOut:     outBuf.String(),
			StdErr:     errBuf.String(),
			StatusCode: -1,
		}, fmt.Errorf("ContainerExecInspect: %w", err)
	}

	return repo.ExecCommandResp{
		StdOut:     outBuf.String(),
		StdErr:     errBuf.String(),
		StatusCode: inspectResp.ExitCode,
	}, nil
}
