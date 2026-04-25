package entity

import "time"

type PipelineReq struct {
	Path      string
	StudentID int
	LabID     int
	CreatedAt time.Time
}

type PipelineResp struct {
	IsSuccess  bool
	ErrMsg     string
	StudentID  int
	LabID      int
	CreatedAt  time.Time
	FinishedAt time.Time
}
