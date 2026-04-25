package cases

import (
	"context"
	"sync"
	"time"

	"github.com/4otis/neurolab-service/internal/entity"
	"go.uber.org/zap"
)

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
	autotestUseCase AutotestUseCase,
) *WorkerPool {
	return &WorkerPool{
		workerCnt:       workerCnt,
		inputCh:         inputCh,
		outputCh:        outputCh,
		wg:              sync.WaitGroup{},
		logger:          logger,
		autotestUseCase: autotestUseCase,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	ctx, wp.cancelFn = context.WithCancel(ctx)

	for i := 0; i < wp.workerCnt; i++ {
		wp.wg.Add(1)
		go func(id int) {
			wp.worker(ctx, id)
			wp.wg.Done()
		}(i)
	}

}

func (wp *WorkerPool) Stop() {

}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	for {
		select {
		case val, ok := <-wp.inputCh:
			if !ok {
				wp.logger.Error("failed to read from channel",
					zap.Int("worker_id", id),
				)
				break
			}
			val.CreatedAt = time.Now()
			resp, err := wp.autotestUseCase.Process(ctx, &val)
			if err != nil {
				// TODO: необходимо реализовать retry механизм внутри Process
				wp.logger.Error("failed to process",
					zap.Int("worker_id", id),
					zap.Any("req", val),
					zap.Error(err),
				)
				break
			}

			if resp != nil {
				processingTime := resp.FinishedAt.Sub(resp.CreatedAt).Milliseconds()
				wp.logger.Debug("solution was processed",
					zap.Int("worker_id", id),
					zap.Any("resp", resp),
					zap.Int("processing_time_ms", int(processingTime)),
				)
				wp.outputCh <- *resp
			}

		case <-ctx.Done():
			wp.logger.Error("canceled by ctx.Done()",
				zap.Int("worker_id", id),
			)
			return
		}
	}
}
