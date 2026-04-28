package awsutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/zap"
)

// NewWebIdentityCredentials returns a credentials provider that assumes roleARN by
// exchanging a web identity token from retriever (a file or GCP metadata server).
// No base AWS credentials are required; the OIDC token authenticates the STS call.
func NewWebIdentityCredentials(
	ctx context.Context,
	roleARN, sessionName, region string,
	retriever stscreds.IdentityTokenRetriever,
	logger *zap.Logger,
) (aws.CredentialsProvider, error) {
	if region == "" {
		region = "us-east-1"
	}
	if sessionName == "" {
		sessionName = "rill-session"
	}
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithLogger(NewAWSLogger(logger)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create STS config for web identity: %w", err)
	}
	provider := stscreds.NewWebIdentityRoleProvider(
		sts.NewFromConfig(cfg), roleARN, retriever,
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = sessionName
		},
	)
	return aws.NewCredentialsCache(provider), nil
}
