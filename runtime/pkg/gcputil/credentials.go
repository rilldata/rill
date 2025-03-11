package gcputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

var ErrNoCredentials = errors.New("empty credentials: set `google_application_credentials` env variable")

func Credentials(ctx context.Context, jsonData string, allowHostAccess bool) (*google.Credentials, error) {
	if jsonData != "" {
		// google_application_credentials is set, use credentials from json string provided by user
		return google.CredentialsFromJSON(ctx, []byte(jsonData), "https://www.googleapis.com/auth/cloud-platform")
	}
	// google_application_credentials is not set
	if allowHostAccess {
		// use host credentials
		creds, err := gcp.DefaultCredentials(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "google: could not find default credentials") {
				return nil, fmt.Errorf("%w: %w", ErrNoCredentials, err)
			}
			return nil, err
		}
		return creds, nil
	}
	return nil, ErrNoCredentials
}

func ProjectID(credentials *google.Credentials) (string, error) {
	projectID := credentials.ProjectID
	if projectID == "" {
		if len(credentials.JSON) == 0 {
			return "", fmt.Errorf("unable to get project ID")
		}
		f := &credentialsFile{}
		if err := json.Unmarshal(credentials.JSON, f); err != nil {
			return "", err
		}

		projectID = f.getProjectID()
	}
	return projectID, nil
}

// credentialsFile is the unmarshalled representation of a credentials file.
type credentialsFile struct {
	Type string `json:"type"`

	// Service Account fields
	ProjectID string `json:"project_id"`

	// External Account fields
	QuotaProjectID string `json:"quota_project_id"`

	// Service account impersonation
	SourceCredentials *credentialsFile `json:"source_credentials"`
}

func (c *credentialsFile) getProjectID() string {
	if c.Type == "impersonated_service_account" {
		return c.SourceCredentials.getProjectID()
	}
	if c.ProjectID != "" {
		return c.ProjectID
	}
	return c.QuotaProjectID
}
