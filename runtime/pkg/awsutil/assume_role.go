package awsutil

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/zap"
)

// NewAssumeRoleCredentials returns credentials for roleARN using sourceCredentials.
// It can be used to chain from credentials obtained through web identity federation.
func NewAssumeRoleCredentials(
	ctx context.Context,
	roleARN, sessionName, externalID, region string,
	sourceCredentials aws.CredentialsProvider,
	logger *zap.Logger,
) (aws.CredentialsProvider, error) {
	if roleARN == "" {
		return nil, fmt.Errorf("cannot assume role: role ARN is empty")
	}
	if sourceCredentials == nil {
		return nil, fmt.Errorf("cannot assume role %q: source credentials are nil", roleARN)
	}
	if sessionName == "" {
		sessionName = "rill-session"
	}

	cfg, err := loadSTSConfig(ctx, region, sourceCredentials, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create STS config for role assumption: %w", err)
	}

	provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), roleARN, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = sessionName
		// Role chaining sessions cannot exceed one hour.
		o.Duration = time.Hour
		if externalID != "" {
			o.ExternalID = &externalID
		}
	})
	return newCredentialsCache(provider), nil
}

func loadSTSConfig(ctx context.Context, region string, credentials aws.CredentialsProvider, logger *zap.Logger) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithCredentialsProvider(credentials),
		config.WithLogger(NewAWSLogger(logger)),
	}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, err
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	return cfg, nil
}
