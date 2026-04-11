package docker

import (
	"context"
	"fmt"

	"github.com/4otis/neurolab-service/internal/port/repo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
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
			target = "app/solutions"
		} else {
			target = "app/scripts"
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

	return resp.ID, nil
}

func (r *DockerRepo) DeleteContainer(ctx context.Context, containerID string) error {
	return nil
}

func (r *DockerRepo) ExecCommand(ctx context.Context, containerID string, cmd string) (repo.ExecCommandResp, error) {
	return repo.ExecCommandResp{}, nil
}
