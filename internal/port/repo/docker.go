package repo

import "context"

type ExecCommandResp struct {
	StdOut     string
	StdErr     string
	StatusCode int
}

type DockerRepo interface {
	CreateContainer(ctx context.Context, dockerImage string) error
	DeleteContainer(ctx context.Context, containerID string) error
	LoadFileToContainer(ctx context.Context, containerID, srcPath, dstPath string) error
	ExecCommand(ctx context.Context, containerID string, cmd string) (ExecCommandResp, error)
}
