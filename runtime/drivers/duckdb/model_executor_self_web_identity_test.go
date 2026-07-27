package duckdb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGenerateSecretSQLWithAWSCredentials(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("test-jwt\n"), 0o600))

	tests := []struct {
		name                  string
		config                map[string]any
		wantAccessKeyID       string
		wantSecretAccessKey   string
		wantSessionToken      string
		wantActions           []string
		wantAssumeSigningKey  string
		wantAssumeSourceToken string
		forbiddenValues       []string
	}{
		{
			name: "static access keys",
			config: map[string]any{
				"region":                "us-east-1",
				"aws_access_key_id":     "STATIC_KEY",
				"aws_secret_access_key": "STATIC_SECRET",
				"aws_access_token":      "STATIC_TOKEN",
			},
			wantAccessKeyID:     "STATIC_KEY",
			wantSecretAccessKey: "STATIC_SECRET",
			wantSessionToken:    "STATIC_TOKEN",
		},
		{
			name: "STS assume role",
			config: map[string]any{
				"region":                "us-east-1",
				"aws_access_key_id":     "BASE_KEY",
				"aws_secret_access_key": "BASE_SECRET",
				"aws_access_token":      "BASE_TOKEN",
				"aws_role_arn":          "arn:aws:iam::210987654321:role/s3-access",
				"aws_role_session_name": "target-session",
				"aws_external_id":       "external-id",
			},
			wantAccessKeyID:       "FINAL_KEY",
			wantSecretAccessKey:   "FINAL_SECRET",
			wantSessionToken:      "FINAL_TOKEN",
			wantActions:           []string{"AssumeRole"},
			wantAssumeSigningKey:  "BASE_KEY",
			wantAssumeSourceToken: "BASE_TOKEN",
			forbiddenValues:       []string{"BASE_KEY", "BASE_SECRET", "BASE_TOKEN"},
		},
		{
			name: "direct WebIdentity",
			config: map[string]any{
				"region":                             "us-east-1",
				"aws_web_identity_token_file":        tokenPath,
				"aws_web_identity_role_arn":          "arn:aws:iam::123456789012:role/web-identity",
				"aws_web_identity_role_session_name": "web-session",
			},
			wantAccessKeyID:     "WEB_KEY",
			wantSecretAccessKey: "WEB_SECRET",
			wantSessionToken:    "WEB_TOKEN",
			wantActions:         []string{"AssumeRoleWithWebIdentity"},
			forbiddenValues:     []string{"test-jwt"},
		},
		{
			name: "WebIdentity then STS assume role",
			config: map[string]any{
				"region":                             "us-east-1",
				"aws_web_identity_token_file":        tokenPath,
				"aws_web_identity_role_arn":          "arn:aws:iam::123456789012:role/web-identity",
				"aws_web_identity_role_session_name": "web-session",
				"aws_role_arn":                       "arn:aws:iam::210987654321:role/s3-access",
				"aws_role_session_name":              "target-session",
				"aws_external_id":                    "external-id",
			},
			wantAccessKeyID:       "FINAL_KEY",
			wantSecretAccessKey:   "FINAL_SECRET",
			wantSessionToken:      "FINAL_TOKEN",
			wantActions:           []string{"AssumeRoleWithWebIdentity", "AssumeRole"},
			wantAssumeSigningKey:  "WEB_KEY",
			wantAssumeSourceToken: "WEB_TOKEN",
			forbiddenValues:       []string{"WEB_KEY", "WEB_SECRET", "WEB_TOKEN", "test-jwt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, requests := newDuckDBTestSTSServer(t)
			t.Setenv("AWS_ENDPOINT_URL_STS", server.URL)

			connector, err := drivers.Open("s3", "", "default", tt.config, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, connector.Close()) })

			opts := &drivers.ModelExecuteOptions{
				ModelExecutorOptions: &drivers.ModelExecutorOptions{
					ModelName: "aws_credentials_model",
					Env: &drivers.ModelEnv{
						AcquireConnector: func(_ context.Context, name string) (drivers.Handle, func(), error) {
							if name != "s3_connector" {
								return nil, nil, fmt.Errorf("unexpected connector %q", name)
							}
							return connector, func() {}, nil
						},
					},
				},
			}

			secretSQL, dropSQL, connectorType, err := generateSecretSQL(
				context.Background(),
				opts,
				"s3_connector",
				"s3://bucket/path",
				map[string]any{},
				zap.NewNop(),
			)
			require.NoError(t, err)
			require.Equal(t, "s3", connectorType)
			require.Contains(t, dropSQL, "DROP SECRET IF EXISTS")
			require.Contains(t, secretSQL, "KEY_ID '"+tt.wantAccessKeyID+"'")
			require.Contains(t, secretSQL, "SECRET '"+tt.wantSecretAccessKey+"'")
			require.Contains(t, secretSQL, "SESSION_TOKEN '"+tt.wantSessionToken+"'")
			require.Contains(t, secretSQL, "REGION 'us-east-1'")
			for _, forbiddenValue := range tt.forbiddenValues {
				require.NotContains(t, secretSQL, forbiddenValue)
			}

			gotRequests := requests.snapshot()
			require.Len(t, gotRequests, len(tt.wantActions))
			for i, wantAction := range tt.wantActions {
				require.Equal(t, wantAction, gotRequests[i].action)
			}
			if tt.wantAssumeSigningKey != "" {
				assumeRequest := gotRequests[len(gotRequests)-1]
				require.Contains(t, assumeRequest.authorization, "Credential="+tt.wantAssumeSigningKey+"/")
				require.Equal(t, tt.wantAssumeSourceToken, assumeRequest.securityToken)
			}
		})
	}
}

type duckDBSTSRequest struct {
	action        string
	authorization string
	securityToken string
}

type duckDBSTSRequests struct {
	mu       sync.Mutex
	requests []duckDBSTSRequest
}

func (r *duckDBSTSRequests) append(request duckDBSTSRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, request)
}

func (r *duckDBSTSRequests) snapshot() []duckDBSTSRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]duckDBSTSRequest(nil), r.requests...)
}

func newDuckDBTestSTSServer(t *testing.T) (*httptest.Server, *duckDBSTSRequests) {
	t.Helper()
	expiration := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	requests := &duckDBSTSRequests{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requests.append(duckDBSTSRequest{
			action:        r.Form.Get("Action"),
			authorization: r.Header.Get("Authorization"),
			securityToken: r.Header.Get("X-Amz-Security-Token"),
		})
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
	return server, requests
}
