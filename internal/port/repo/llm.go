package repo

import "context"

type LLMRepo interface {
	Generate(ctx context.Context, sysP, usrP string) (rawResp string, err error)
}
