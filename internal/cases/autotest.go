package cases

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/4otis/neurolab-service/internal/entity"
	"github.com/4otis/neurolab-service/internal/port/repo"
	"go.uber.org/zap"
)

var _ AutotestUseCase = (*AutotestUseCaseImpl)(nil)

type AutotestUseCase interface {
	Process(ctx context.Context, req *entity.PipelineReq) (*entity.PipelineResp, error)
}

type AutotestUseCaseImpl struct {
	logger       *zap.Logger
	labRepo      repo.LabRepo
	dockerRepo   repo.DockerRepo
	solutionsDir string
	scriptsDir   string
	defaultImage string
	dockerConfig map[string]string
}

func NewAutotestUseCase(
	logger *zap.Logger,
	labRepo repo.LabRepo,
	dockerRepo repo.DockerRepo,
	solutionsDir string,
	scriptsDir string,
) *AutotestUseCaseImpl {
	//TODO: вынести в отдельный конфигурационный файл .env или .yaml
	cfg := map[string]string{
		"python": "checker-python:1.0",
		"golang": "checker-golang:1.0",
		"cpp":    "checker-cpp:1.0",
	}

	return &AutotestUseCaseImpl{
		logger:       logger,
		labRepo:      labRepo,
		dockerRepo:   dockerRepo,
		solutionsDir: solutionsDir,
		scriptsDir:   scriptsDir,
		defaultImage: "checker-python:1.0", // TODO: подумать о надобности
		dockerConfig: cfg,
	}
}

func (uc *AutotestUseCaseImpl) Process(ctx context.Context, req *entity.PipelineReq) (pipelineResp *entity.PipelineResp, err error) {
	pipelineResp = &entity.PipelineResp{
		IsSuccess: true,
		StudentID: req.StudentID,
		LabID:     req.LabID,
		CreatedAt: req.CreatedAt,
	}

	lab, err := uc.labRepo.Read(ctx, req.LabID)
	if err != nil {
		uc.logger.Error("failed to read pipeline execute plan",
			zap.Int("lab_id", req.LabID),
			zap.Error(err),
		)
		return pipelineResp, fmt.Errorf("failed to read pipeline execute plan:%w", err)
	}

	scripts := strings.Split(lab.Pipeline, "->")
	image := uc.dockerConfig[lab.Lang]

	mnts := make([]*string, 0, len(scripts)+1)

	absSolutionPath, err := filepath.Abs(req.Path)
	if err != nil {
		uc.logger.Error("failed to get absolute path for solution", zap.Error(err))
		return pipelineResp, fmt.Errorf("invalid solution path: %w", err)
	}
	mnts = append(mnts, &absSolutionPath)

	absScriptsPath, err := filepath.Abs(uc.scriptsDir)
	if err != nil {
		uc.logger.Error("failed to get absolute path for scripts folder", zap.Error(err))
		return pipelineResp, fmt.Errorf("invalid scripts path: %w", err)
	}
	mnts = append(mnts, &absScriptsPath)

	// for _, s := range scripts {
	// 	scriptFullPath := filepath.Join(uc.scriptsDir, s)
	// 	absScriptPath, err := filepath.Abs(scriptFullPath)
	// 	if err != nil {
	// 		uc.logger.Error("failed to get absolute path for script", zap.String("script", s), zap.Error(err))
	// 		return pipelineResp, fmt.Errorf("invalid script path: %w", err)
	// 	}
	// 	mnts = append(mnts, &absScriptPath)
	// }

	uc.logger.Debug("show mounts", zap.Any("mnts", mnts))

	containerID, err := uc.dockerRepo.CreateContainer(ctx, image, mnts)
	if err != nil {
		uc.logger.Error("failed to create container",
			zap.String("image", image),
			zap.Error(err),
		)
		return pipelineResp, fmt.Errorf("failed to create container:%w", err)
	}
	defer func() {
		if rmErr := uc.dockerRepo.DeleteContainer(ctx, containerID); rmErr != nil {
			uc.logger.Error("failed to remove container", zap.String("container_id", containerID), zap.Error(rmErr))
		}
	}()

	for _, s := range scripts {
		cmd := []string{fmt.Sprintf("/app/scripts/%s/run.sh", s)}
		resp, err := uc.dockerRepo.ExecCommand(ctx, containerID, cmd)
		if err != nil {
			uc.logger.Error("failed to exec command",
				zap.String("cmd", cmd[0]),
				zap.Error(err),
			)
			return pipelineResp, err
		}

		if resp.StatusCode != 0 {
			uc.logger.Debug("bad status code",
				zap.String("container_id", containerID),
				zap.String("script", s),
				zap.Any("resp", resp),
			)
			pipelineResp.IsSuccess = false
			break
		} else {
			uc.logger.Debug("script passed",
				zap.String("container_id", containerID),
				zap.String("script", s),
				zap.Any("resp", resp),
			)
		}
	}

	pipelineResp.FinishedAt = time.Now()
	return pipelineResp, nil
}
