package duckdbreplicator

import (
	"context"
	"errors"
	"strings"

	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

var ErrNoCredentials = errors.New("empty credentials: set `google_application_credentials` env variable")

func newClient(ctx context.Context, jsonData string, allowHostAccess bool) (*gcp.HTTPClient, error) {
	creds, err := credentials(ctx, jsonData, allowHostAccess)
	if err != nil {
		if !errors.Is(err, ErrNoCredentials) {
			return nil, err
		}

		// no credentials set, we try with a anonymous client in case user is trying to access public buckets
		return gcp.NewAnonymousHTTPClient(gcp.DefaultTransport()), nil
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}

func credentials(ctx context.Context, jsonData string, allowHostAccess bool) (*google.Credentials, error) {
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
				return nil, ErrNoCredentials
			}

			return nil, err
		}
		return creds, nil
	}
	return nil, ErrNoCredentials
}
