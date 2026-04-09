package cases

import (
	"context"
	"sync"

	"github.com/4otis/neurolab-service/internal/entity"
	"go.uber.org/zap"
)

// type Worker interface {
// 	Process(ctx context.Context) error
// }

type WorkerPool struct {
	workerCnt       int
	inputCh         <-chan entity.PipelineReq
	outputCh        chan<- entity.PipelineResp
	cancelFn        context.CancelFunc
	wg              sync.WaitGroup
	logger          *zap.Logger
	autotestUseCase AutotestUseCase
}

func NewWorkerPool(
	workerCnt int,
	inputCh <-chan entity.PipelineReq,
	outputCh chan<- entity.PipelineResp,
	logger *zap.Logger,
) *WorkerPool {
	return &WorkerPool{
		workerCnt: workerCnt,
		inputCh:   inputCh,
		outputCh:  outputCh,
		wg:        sync.WaitGroup{},
		logger:    logger,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	ctx, wp.cancelFn = context.WithCancel(ctx)

	for i := 0; i < wp.workerCnt; i++ {
		go func(id int) {
			wp.wg.Add(1)
			wp.worker(ctx, id)
			wp.wg.Done()
		}(i)
	}

}

func (wp *WorkerPool) Stop() {

}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	for {
		// select {
		// case val, ok := <-wp.inputCh:
		// 	if !ok {
		// 		wp.logger.Error("failed to read from channel",
		// 			zap.Int("worker_id", id),
		// 		)
		// 	}
		// 	wp.autotestUseCase.Process(ctx, val)
		// case <-ctx.Done:
		// }
	}
}
