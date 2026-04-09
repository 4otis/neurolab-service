package postgres

import (
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
