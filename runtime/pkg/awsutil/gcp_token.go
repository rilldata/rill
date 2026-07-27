package awsutil

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
)

const gcpMetadataTokenTimeout = 5 * time.Second

// GCPMetadataTokenRetriever fetches a Google-signed OIDC JWT from the GCP instance
// metadata server for exchange with AWS STS via AssumeRoleWithWebIdentity.
type GCPMetadataTokenRetriever struct {
	Audience string
	client   *metadata.Client
}

// GetIdentityToken implements stscreds.IdentityTokenRetriever.
func (r GCPMetadataTokenRetriever) GetIdentityToken() ([]byte, error) {
	client := r.client
	if client == nil {
		client = metadata.NewClient(nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gcpMetadataTokenTimeout)
	defer cancel()

	path := "instance/service-accounts/default/identity?audience=" + url.QueryEscape(r.Audience) + "&format=full"
	token, err := client.GetWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve identity token from GCP metadata server: %w", err)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("GCP metadata server returned an empty identity token")
	}
	return []byte(token), nil
}
