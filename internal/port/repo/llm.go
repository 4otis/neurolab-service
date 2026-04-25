package repo

import "context"

type GenerateRequest struct {
	Prompt  string `json:"prompt"`
	Session string `json:"session_id,omitempty"`
}

type GenerateResponse struct {
	Content   string `json:"content"`
	SessionID string `json:"session_id"`
	ValidJSON bool   `json:"valid_json"`
}

type LLMRepo interface {
	Generate(ctx context.Context, sysP, usrP string) (rawResp string, err error)
}
