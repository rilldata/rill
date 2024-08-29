package jobs

import (
	"context"
)

type Client interface {
	Close(ctx context.Context) error
	Work(ctx context.Context) error
	CancelJob(ctx context.Context, jobID int64) error

	// NOTE: Add new job trigger functions here
	ResetAllDeployments(ctx context.Context) (*InsertResult, error)
}

type InsertResult struct {
	ID        int64
	Duplicate bool
}
