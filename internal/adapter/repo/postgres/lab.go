package postgres

import (
	"context"

	"github.com/4otis/neurolab-service/internal/entity"
	"github.com/4otis/neurolab-service/internal/port/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repo.LabRepo = (*LabRepo)(nil)

type LabRepo struct {
	pool *pgxpool.Pool
}

func NewLabRepo(pool *pgxpool.Pool) *LabRepo {
	return &LabRepo{pool: pool}
}

func (r *LabRepo) Read(ctx context.Context, id int) (*entity.Lab, error) {
	return &entity.Lab{
		ID:       id,
		Name:     "lab name",
		Task:     "do some task",
		Lang:     "python",
		Pipeline: "plagiarism_simple->python_pep8->python_lab_01_task_01",
	}, nil
}
