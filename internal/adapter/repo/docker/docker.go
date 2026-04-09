package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerRepo struct {
	dockerClient *client.Client
}

func NewDockerRepo(dockerClient *client.Client) *DockerRepo {
	return &DockerRepo{dockerClient: dockerClient}
}

func (r *DockerRepo) CreateContainer(ctx context.Context, imageName string) (respID string, err error) {
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   []string{"sleep", "infinity"},
	}

	hostConfig := &container.HostConfig{
		AutoRemove: false,
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

func (r *DockerRepo) LoadFileToContainer(ctx context.Context, containerID, filepath string) error {
	// tar
	// r.dockerClient.CopyToContainer()

	return nil
}
