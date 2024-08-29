package jobs

import "context"

type Client interface {
	Close(ctx context.Context) error
	Work(ctx context.Context) error
	//  NOTE: Add new job trigger functions here
	ResetAllDeployments(ctx context.Context) error
}
