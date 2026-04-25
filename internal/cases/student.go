package cases

import "context"

type StudentUseCase interface {
	GetAvailableLabs(ctx context.Context, studentId int) (struct{}, error)
}
