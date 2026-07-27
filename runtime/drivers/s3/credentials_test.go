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
	tests := []struct {
		name             string
		config           *ConfigProperties
		wantAccessKeyID  string
		wantActions      []string
		wantWebSession   string
		wantTargetRole   bool
		wantTargetARN    string
		wantTargetHeader bool
	}{
		{
			name: "legacy direct role",
			config: &ConfigProperties{
				RoleARN:         testWebRoleARN,
				RoleSessionName: "legacy-session",
			},
			wantAccessKeyID: "WEB_KEY",
			wantActions:     []string{"AssumeRoleWithWebIdentity"},
			wantWebSession:  "legacy-session",
		},
		{
			name: "explicit direct role",
			config: &ConfigProperties{
				WebIdentityRoleARN:         testWebRoleARN,
				WebIdentityRoleSessionName: "web-session",
			},
			wantAccessKeyID: "WEB_KEY",
			wantActions:     []string{"AssumeRoleWithWebIdentity"},
			wantWebSession:  "web-session",
		},
		{
			name: "chained target role",
			config: &ConfigProperties{
				WebIdentityRoleARN:         testWebRoleARN,
				WebIdentityRoleSessionName: "web-session",
				RoleARN:                    testTargetRoleARN,
				RoleSessionName:            "target-session",
				ExternalID:                 "external-id",
			},
			wantAccessKeyID:  "FINAL_KEY",
			wantActions:      []string{"AssumeRoleWithWebIdentity", "AssumeRole"},
			wantWebSession:   "web-session",
			wantTargetRole:   true,
			wantTargetARN:    testTargetRoleARN,
			wantTargetHeader: true,
		},
		{
			name: "explicit same role still chains",
			config: &ConfigProperties{
				WebIdentityRoleARN:         testWebRoleARN,
				WebIdentityRoleSessionName: "web-session",
				RoleARN:                    testWebRoleARN,
				RoleSessionName:            "target-session",
				ExternalID:                 "external-id",
			},
			wantAccessKeyID:  "FINAL_KEY",
			wantActions:      []string{"AssumeRoleWithWebIdentity", "AssumeRole"},
			wantWebSession:   "web-session",
			wantTargetRole:   true,
			wantTargetARN:    testWebRoleARN,
			wantTargetHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, requests := newTestSTSServer(t)
			t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)

			tokenPath := filepath.Join(t.TempDir(), "token")
			require.NoError(t, os.WriteFile(tokenPath, []byte("test-jwt\n"), 0o600))
			tt.config.WebIdentityTokenFile = tokenPath
			tt.config.Region = "us-east-1"

			cfg, err := GetConfigWithTemporaryCredentials(context.Background(), tt.config, zap.NewNop())
			require.NoError(t, err)
			require.Equal(t, tt.wantAccessKeyID, cfg.AccessKeyID)
			if tt.wantTargetRole {
				require.Equal(t, "FINAL_SECRET", cfg.SecretAccessKey)
				require.Equal(t, "FINAL_TOKEN", cfg.SessionToken)
			} else {
				require.Equal(t, "WEB_SECRET", cfg.SecretAccessKey)
				require.Equal(t, "WEB_TOKEN", cfg.SessionToken)
			}
			requireCredentialSourcesCleared(t, cfg)

			gotRequests := requests.snapshot()
			require.Len(t, gotRequests, len(tt.wantActions))
			for i, wantAction := range tt.wantActions {
				require.Equal(t, wantAction, gotRequests[i].form.Get("Action"))
			}

			webRequest := gotRequests[0]
			require.Equal(t, testWebRoleARN, webRequest.form.Get("RoleArn"))
			require.Equal(t, tt.wantWebSession, webRequest.form.Get("RoleSessionName"))
			require.Equal(t, "test-jwt", webRequest.form.Get("WebIdentityToken"))
			require.Equal(t, "3600", webRequest.form.Get("DurationSeconds"))
			require.Empty(t, webRequest.header.Get("Authorization"))

			if tt.wantTargetHeader {
				targetRequest := gotRequests[1]
				require.Equal(t, tt.wantTargetARN, targetRequest.form.Get("RoleArn"))
				require.Equal(t, "target-session", targetRequest.form.Get("RoleSessionName"))
				require.Equal(t, "external-id", targetRequest.form.Get("ExternalId"))
				require.Equal(t, "3600", targetRequest.form.Get("DurationSeconds"))
				require.Contains(t, targetRequest.header.Get("Authorization"), "Credential=WEB_KEY/")
				require.Equal(t, "WEB_TOKEN", targetRequest.header.Get("X-Amz-Security-Token"))
			}
		})
	}
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
	require.Contains(t, gotRequests[0].header.Get("Authorization"), "/eu-west-1/sts/")
	require.Equal(t, "BASE_TOKEN", gotRequests[0].header.Get("X-Amz-Security-Token"))
}

func TestWebIdentityConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConfigProperties
		wantErr string
	}{
		{
			name: "token without role",
			config: &ConfigProperties{
				WebIdentityTokenFile: "/token",
			},
			wantErr: "require aws_web_identity_role_arn or aws_role_arn",
		},
		{
			name: "web role without token",
			config: &ConfigProperties{
				WebIdentityRoleARN: testWebRoleARN,
			},
			wantErr: "requires aws_web_identity_token_file",
		},
		{
			name: "external ID without target role",
			config: &ConfigProperties{
				RoleARN:              testWebRoleARN,
				ExternalID:           "external-id",
				WebIdentityTokenFile: "/token",
				Region:               "us-east-1",
			},
			wantErr: "requires a separate aws_role_arn target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newCredentialsProvider(context.Background(), tt.config, zap.NewNop())
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func requireCredentialSourcesCleared(t *testing.T, cfg *ConfigProperties) {
	t.Helper()
	require.Empty(t, cfg.RoleARN)
	require.Empty(t, cfg.RoleSessionName)
	require.Empty(t, cfg.ExternalID)
	require.Empty(t, cfg.WebIdentityTokenFile)
	require.Empty(t, cfg.WebIdentityRoleARN)
	require.Empty(t, cfg.WebIdentityRoleSessionName)
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
