package awsutil

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"go.uber.org/zap"
)

const credentialsExpiryWindow = 5 * time.Minute

// NewWebIdentityCredentials returns a credentials provider that assumes roleARN by
// exchanging a web identity token from retriever (a file or GCP metadata server).
// No base AWS credentials are required; the OIDC token authenticates the STS call.
func NewWebIdentityCredentials(
	ctx context.Context,
	roleARN, sessionName, region string,
	retriever stscreds.IdentityTokenRetriever,
	logger *zap.Logger,
) (aws.CredentialsProvider, error) {
	if roleARN == "" {
		return nil, fmt.Errorf("cannot assume role with web identity: role ARN is empty")
	}
	if retriever == nil {
		return nil, fmt.Errorf("cannot assume role %q with web identity: token retriever is nil", roleARN)
	}
	if sessionName == "" {
		sessionName = "rill-session"
	}
	cfg, err := loadSTSConfig(ctx, region, aws.AnonymousCredentials{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create STS config for web identity: %w", err)
	}
	provider := stscreds.NewWebIdentityRoleProvider(
		sts.NewFromConfig(cfg), roleARN, trimmingIdentityTokenRetriever{retriever: retriever},
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = sessionName
			o.Duration = time.Hour
		},
	)
	return newCredentialsCache(provider), nil
}

type trimmingIdentityTokenRetriever struct {
	retriever stscreds.IdentityTokenRetriever
}

func (r trimmingIdentityTokenRetriever) GetIdentityToken() ([]byte, error) {
	token, err := r.retriever.GetIdentityToken()
	if err != nil {
		return nil, err
	}
	token = bytes.TrimSpace(token)
	if len(token) == 0 {
		return nil, fmt.Errorf("identity token is empty")
	}
	return token, nil
}

func newCredentialsCache(provider aws.CredentialsProvider) *aws.CredentialsCache {
	return aws.NewCredentialsCache(provider, func(o *aws.CredentialsCacheOptions) {
		o.ExpiryWindow = credentialsExpiryWindow
		o.ExpiryWindowJitterFrac = 0.2
	})
}
