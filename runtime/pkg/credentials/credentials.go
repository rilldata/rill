package credentials

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

// AssumeRoleOptions specifies the parameters for assuming an AWS role.
type AssumeRoleOptions struct {
	RoleARN         string
	RoleSessionName string
	DurationSeconds *int32 // Optional: defaults to 1 hour (3600s) if nil. AWS min 900, max depends on role.
}

// AssumeRole assumes an AWS IAM role and returns temporary credentials.
// It uses the default AWS credential chain from the environment to make the AssumeRole call.
func AssumeRole(ctx context.Context, opts AssumeRoleOptions) (*types.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("sts: failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	roleSessionName := opts.RoleSessionName
	if roleSessionName == "" {
		roleSessionName = "rill-session"
	}

	input := &sts.AssumeRoleInput{
		RoleArn:         &opts.RoleARN,
		RoleSessionName: &roleSessionName,
	}
	if opts.DurationSeconds != nil {
		input.DurationSeconds = opts.DurationSeconds
	}

	result, err := stsClient.AssumeRole(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("sts: failed to assume role %s: %w", opts.RoleARN, err)
	}

	if result.Credentials == nil {
		return nil, fmt.Errorf("sts: AssumeRole call for %s returned nil credentials", opts.RoleARN)
	}

	return result.Credentials, nil
}
