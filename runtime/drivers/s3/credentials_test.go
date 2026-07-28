package s3

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testWebRoleARN    = "arn:aws:iam::123456789012:role/web-identity"
	testTargetRoleARN = "arn:aws:iam::210987654321:role/s3-access"
)

type recordedSTSRequest struct {
	form   url.Values
	header http.Header
}

func TestGetConfigWithTemporaryCredentialsWebIdentity(t *testing.T) {
	server, requests := newTestSTSServer(t)
	t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)

	tokenPath := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("test-jwt\n"), 0o600))

	cfg, err := GetConfigWithTemporaryCredentials(context.Background(), &ConfigProperties{
		RoleARN:              testWebRoleARN,
		RoleSessionName:      "web-session",
		WebIdentityTokenFile: tokenPath,
		Region:               "us-east-1",
	}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "WEB_KEY", cfg.AccessKeyID)
	require.Equal(t, "WEB_SECRET", cfg.SecretAccessKey)
	require.Equal(t, "WEB_TOKEN", cfg.SessionToken)
	requireCredentialSourcesCleared(t, cfg)

	gotRequests := requests.snapshot()
	require.Len(t, gotRequests, 1)
	webRequest := gotRequests[0]
	require.Equal(t, "AssumeRoleWithWebIdentity", webRequest.form.Get("Action"))
	require.Equal(t, testWebRoleARN, webRequest.form.Get("RoleArn"))
	require.Equal(t, "web-session", webRequest.form.Get("RoleSessionName"))
	require.Equal(t, "test-jwt", webRequest.form.Get("WebIdentityToken"))
	require.Equal(t, "3600", webRequest.form.Get("DurationSeconds"))
	require.Empty(t, webRequest.header.Get("Authorization"))
}

func TestGetConfigWithTemporaryCredentialsAssumeRole(t *testing.T) {
	server, requests := newTestSTSServer(t)
	t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)
	t.Setenv("AWS_REGION", "eu-west-1")

	cfg, err := GetConfigWithTemporaryCredentials(context.Background(), &ConfigProperties{
		AccessKeyID:     "BASE_KEY",
		SecretAccessKey: "BASE_SECRET",
		SessionToken:    "BASE_TOKEN",
		RoleARN:         testTargetRoleARN,
		RoleSessionName: "target-session",
		ExternalID:      "external-id",
	}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "FINAL_KEY", cfg.AccessKeyID)
	require.Equal(t, "FINAL_SECRET", cfg.SecretAccessKey)
	require.Equal(t, "FINAL_TOKEN", cfg.SessionToken)
	requireCredentialSourcesCleared(t, cfg)

	gotRequests := requests.snapshot()
	require.Len(t, gotRequests, 1)
	require.Equal(t, "AssumeRole", gotRequests[0].form.Get("Action"))
	require.Equal(t, testTargetRoleARN, gotRequests[0].form.Get("RoleArn"))
	require.Contains(t, gotRequests[0].header.Get("Authorization"), "Credential=BASE_KEY/")
	require.Contains(t, gotRequests[0].header.Get("Authorization"), "/us-east-1/sts/")
	require.Equal(t, "BASE_TOKEN", gotRequests[0].header.Get("X-Amz-Security-Token"))
}

func requireCredentialSourcesCleared(t *testing.T, cfg *ConfigProperties) {
	t.Helper()
	require.Empty(t, cfg.RoleARN)
	require.Empty(t, cfg.RoleSessionName)
	require.Empty(t, cfg.ExternalID)
	require.Empty(t, cfg.WebIdentityTokenFile)
}

type recordedSTSRequests struct {
	mu       sync.Mutex
	requests []recordedSTSRequest
}

func (r *recordedSTSRequests) append(request recordedSTSRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, request)
}

func (r *recordedSTSRequests) snapshot() []recordedSTSRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedSTSRequest(nil), r.requests...)
}

func newTestSTSServer(t *testing.T) (*httptest.Server, *recordedSTSRequests) {
	t.Helper()
	records := &recordedSTSRequests{}
	expiration := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		records.append(recordedSTSRequest{form: cloneValues(r.Form), header: r.Header.Clone()})
		w.Header().Set("Content-Type", "text/xml")

		switch r.Form.Get("Action") {
		case "AssumeRoleWithWebIdentity":
			_, _ = fmt.Fprintf(w, `<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/"><AssumeRoleWithWebIdentityResult><Credentials><AccessKeyId>WEB_KEY</AccessKeyId><SecretAccessKey>WEB_SECRET</SecretAccessKey><SessionToken>WEB_TOKEN</SessionToken><Expiration>%s</Expiration></Credentials></AssumeRoleWithWebIdentityResult><ResponseMetadata><RequestId>web-request</RequestId></ResponseMetadata></AssumeRoleWithWebIdentityResponse>`, expiration)
		case "AssumeRole":
			_, _ = fmt.Fprintf(w, `<AssumeRoleResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/"><AssumeRoleResult><Credentials><AccessKeyId>FINAL_KEY</AccessKeyId><SecretAccessKey>FINAL_SECRET</SecretAccessKey><SessionToken>FINAL_TOKEN</SessionToken><Expiration>%s</Expiration></Credentials></AssumeRoleResult><ResponseMetadata><RequestId>role-request</RequestId></ResponseMetadata></AssumeRoleResponse>`, expiration)
		default:
			http.Error(w, "unexpected action "+r.Form.Get("Action"), http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)
	return server, records
}

func cloneValues(values url.Values) url.Values {
	clone := make(url.Values, len(values))
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}
