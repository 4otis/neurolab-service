package repo

import "context"

type ExecCommandResp struct {
	StdOut     string
	StdErr     string
	StatusCode int
}

type DockerRepo interface {
	CreateContainer(ctx context.Context, imageName string, mnts []*string) (string, error)
	DeleteContainer(ctx context.Context, containerID string) error
	ExecCommand(ctx context.Context, containerID string, cmd string) (ExecCommandResp, error)
}
