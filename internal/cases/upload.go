package cases

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/4otis/neurolab-service/internal/entity"
	"go.uber.org/zap"
)

var _ UploadUseCase = (*UploadUseCaseImpl)(nil)

type UploadUseCase interface {
	UploadSolution(ctx context.Context, studentID, labID int, src io.Reader) error
	UploadScript(ctx context.Context, scriptName string, src io.Reader) error
}

type UploadUseCaseImpl struct {
	logger       *zap.Logger
	pipelineCh   chan entity.PipelineReq
	solutionsDir string
	scriptsDir   string
}

func NewUploadUseCase(
	logger *zap.Logger,
	output chan entity.PipelineReq,
	solutionsDir string,
	scriptsDir string,

) *UploadUseCaseImpl {
	return &UploadUseCaseImpl{
		logger:       logger,
		pipelineCh:   output,
		solutionsDir: solutionsDir,
		scriptsDir:   scriptsDir,
	}
}

func (uc *UploadUseCaseImpl) UploadSolution(ctx context.Context, studentID, labID int, src io.Reader) (err error) {
	uniq := time.Now().UnixNano()
	tmpDir := filepath.Join(uc.solutionsDir, fmt.Sprintf("%d_%d_%d", studentID, labID, uniq))
	os.RemoveAll(tmpDir)
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tmpDir)
		}
	}()

	archivePath := filepath.Join(tmpDir, "archive.zip")
	dst, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive file: %w", err)
	}
	defer dst.Close()

	if err = uc.copyWithContext(ctx, dst, src); err != nil {
		return fmt.Errorf("copy archive content: %w", err)
	}

	req := entity.PipelineReq{
		Path:      tmpDir,
		StudentID: studentID,
		LabID:     labID,
	}

	select {
	case uc.pipelineCh <- req:
		uc.logger.Debug("sending req to pipeline channel",
			zap.Any("pipilineReq", req),
		)
	case <-ctx.Done():
		return ctx.Err()
	}

	uc.logger.Info("solution was uploaded",
		zap.String("workdir", tmpDir),
		zap.Int("student_id", studentID),
		zap.Int("lab_id", labID),
	)

	return nil
}

func (uc *UploadUseCaseImpl) UploadScript(ctx context.Context, scriptName string, src io.Reader) (err error) {
	tmpDir := filepath.Join(uc.scriptsDir, scriptName)
	os.RemoveAll(tmpDir)
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tmpDir)
		}
	}()

	archivePath := filepath.Join(tmpDir, "archive.zip")
	dst, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive file: %w", err)
	}
	defer dst.Close()

	if err = uc.copyWithContext(ctx, dst, src); err != nil {
		return fmt.Errorf("copy archive content: %w", err)
	}

	uc.logger.Info("script was uploaded",
		zap.String("workdir", tmpDir),
		zap.String("script_name", scriptName),
	)

	return nil
}

type ContextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *ContextReader) Read(p []byte) (n int, err error) {
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
		return cr.r.Read(p)
	}
}

func (uc *UploadUseCaseImpl) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	ctxReader := &ContextReader{
		ctx: ctx,
		r:   src,
	}

	_, err := io.Copy(dst, ctxReader)

	return err
}
