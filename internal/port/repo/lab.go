package repo

import (
	"context"

	"github.com/4otis/neurolab-service/internal/entity"
)

type LabRepo interface {
	Read(ctx context.Context, id int) (*entity.Lab, error)
}
