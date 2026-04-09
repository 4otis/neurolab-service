package cases

import (
	"context"

	"github.com/4otis/neurolab-service/internal/entity"
	"github.com/4otis/neurolab-service/internal/port/repo"
	"go.uber.org/zap"
)

var _ AutotestUseCase = (*AutotestUseCaseImpl)(nil)

type AutotestUseCase interface {
	Process(ctx context.Context, req entity.PipelineReq) (entity.PipelineResp, error)
}

func NewAutotestUseCase(
	logger *zap.Logger,
	labRepo repo.LabRepo,
	dockerRepo repo.DockerRepo,
) *AutotestUseCaseImpl {
	return &AutotestUseCaseImpl{
		logger:     logger,
		labRepo:    labRepo,
		dockerRepo: dockerRepo,
	}
}

type AutotestUseCaseImpl struct {
	logger     *zap.Logger
	labRepo    repo.LabRepo
	dockerRepo repo.DockerRepo
}

func (uc *AutotestUseCaseImpl) Process(ctx context.Context, req entity.PipelineReq) (entity.PipelineResp, error) {
	//TODO: GET ScriptExecutePlan ("step1->step2->step3")

	//TODO: UpContainer

	//TODO: MountSolution (cases.containerUseCase? или это repo.ContainerRepo)

	// scripts := []string{"script1", "script2", "plagiarism_v1"}
	// for _, s := range scripts {
	// 	//TODO: MountScript()

	// 	//TODO: ExecScript()
	// 	//TODO: Analyze status code
	// 	//TODO: Accumulating data for PipelineResp

	// }

	return entity.PipelineResp{}, nil
}
