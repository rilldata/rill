package awsutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

//nolint:gosec // not a credential; this is the well-known GCP metadata server URL
const gcpMetadataTokenURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity"

// GCPMetadataTokenRetriever fetches a Google-signed OIDC JWT from the GCP instance
// metadata server for exchange with AWS STS via AssumeRoleWithWebIdentity.
type GCPMetadataTokenRetriever struct {
	Audience string
}

// GetIdentityToken implements stscreds.IdentityTokenRetriever.
func (r GCPMetadataTokenRetriever) GetIdentityToken() ([]byte, error) {
	reqURL := gcpMetadataTokenURL + "?audience=" + url.QueryEscape(r.Audience) + "&format=full"
	req, err := http.NewRequest(http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach GCP metadata server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GCP metadata server returned status %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(token), nil
}
