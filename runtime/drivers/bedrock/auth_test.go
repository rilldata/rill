package bedrock

import (
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuth(t *testing.T) {
	cases := []struct {
		name     string
		config   map[string]any
		envToken string
		// wantAuth is the expected prefix of the Authorization header, or empty if Open should fail.
		wantAuth string
		wantErr  string
	}{
		{
			name:     "api key",
			config:   map[string]any{"api_key": "configured-key"},
			wantAuth: "Bearer configured-key",
		},
		{
			name:     "api key takes precedence over host token",
			config:   map[string]any{"api_key": "configured-key", "allow_host_access": true},
			envToken: "host-key",
			wantAuth: "Bearer configured-key",
		},
		{
			name:     "access keys",
			config:   map[string]any{"aws_access_key_id": "AKIATEST", "aws_secret_access_key": "secret"},
			wantAuth: "AWS4-HMAC-SHA256 Credential=AKIATEST/",
		},
		{
			name:     "access keys take precedence over host token",
			config:   map[string]any{"aws_access_key_id": "AKIATEST", "aws_secret_access_key": "secret", "allow_host_access": true},
			envToken: "host-key",
			wantAuth: "AWS4-HMAC-SHA256 Credential=AKIATEST/",
		},
		{
			name:     "host token is used with host access",
			config:   map[string]any{"allow_host_access": true},
			envToken: "host-key",
			wantAuth: "Bearer host-key",
		},
		{
			name:     "host token is ignored without host access",
			config:   map[string]any{},
			envToken: "host-key",
			wantErr:  "either api_key or aws_access_key_id and aws_secret_access_key are required",
		},
		{
			name:    "api key cannot be combined with access keys",
			config:  map[string]any{"api_key": "configured-key", "aws_access_key_id": "AKIATEST", "aws_secret_access_key": "secret"},
			wantErr: "api_key cannot be combined",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeBedrock(t)
			t.Setenv("AWS_BEARER_TOKEN_BEDROCK", tc.envToken)

			cfg := map[string]any{"model": "test-model", "region": "us-east-1"}
			for k, v := range tc.config {
				cfg[k] = v
			}
			h, err := driver{}.Open(t.Context(), "", "default", cfg, nil, nil, zap.NewNop())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			ai, ok := h.AsAI("default")
			require.True(t, ok)
			res, err := ai.Complete(t.Context(), &drivers.CompleteOptions{
				Messages: []*aiv1.CompletionMessage{{Role: "user", Content: []*aiv1.ContentBlock{text("hi")}}},
			})
			require.NoError(t, err)
			require.True(t, strings.HasPrefix(fake.auth, tc.wantAuth), "got Authorization header %q", fake.auth)

			require.Equal(t, "bedrock", res.Provider)
			require.Equal(t, 10, res.InputTokens)
			require.Equal(t, 4, res.CachedInputTokens)
			require.Equal(t, 3, res.OutputTokens)
			require.Len(t, res.Message.Content, 2)
			call := res.Message.Content[1].GetToolCall()
			require.Equal(t, "call_1", call.Id)
			require.Equal(t, map[string]any{"sql": "SELECT 1", "limit": float64(5)}, call.Input.AsMap())
		})
	}
}

func TestPing(t *testing.T) {
	// Ping skips the STS check when the client authenticates with an API key,
	// including one loaded from the host, in which case there may be no AWS credentials to verify.
	cases := []struct {
		name     string
		config   map[string]any
		envToken string
	}{
		{
			name:   "api key",
			config: map[string]any{"api_key": "configured-key"},
		},
		{
			name:     "host token",
			config:   map[string]any{"allow_host_access": true},
			envToken: "host-key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Point STS at a closed server, so the test fails if Ping calls it.
			newFakeBedrock(t)
			t.Setenv("AWS_ENDPOINT_URL_STS", "https://127.0.0.1:1")
			t.Setenv("AWS_BEARER_TOKEN_BEDROCK", tc.envToken)

			cfg := map[string]any{"model": "test-model", "region": "us-east-1"}
			for k, v := range tc.config {
				cfg[k] = v
			}
			h, err := driver{}.Open(t.Context(), "", "default", cfg, nil, nil, zap.NewNop())
			require.NoError(t, err)
			require.NoError(t, h.Ping(t.Context()))
		})
	}

	t.Run("access keys are verified with STS", func(t *testing.T) {
		newFakeBedrock(t)
		t.Setenv("AWS_ENDPOINT_URL_STS", "https://127.0.0.1:1")
		h, err := driver{}.Open(t.Context(), "", "default", map[string]any{
			"model":                 "test-model",
			"region":                "us-east-1",
			"aws_access_key_id":     "AKIATEST",
			"aws_secret_access_key": "secret",
		}, nil, nil, zap.NewNop())
		require.NoError(t, err)
		require.ErrorContains(t, h.Ping(t.Context()), "failed to verify AWS credentials")
	})
}

func TestInferenceConfig(t *testing.T) {
	// Inference parameters are only sent if configured, since some models (e.g. reasoning models) reject temperature.
	cases := []struct {
		name   string
		config map[string]any
		want   map[string]any
	}{
		{
			name:   "defaults",
			config: map[string]any{},
			want:   nil,
		},
		{
			name:   "configured",
			config: map[string]any{"temperature": 0.5, "max_tokens": 1000},
			want:   map[string]any{"temperature": 0.5, "maxTokens": float64(1000)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeBedrock(t)

			cfg := map[string]any{"model": "test-model", "region": "us-east-1", "api_key": "key"}
			for k, v := range tc.config {
				cfg[k] = v
			}
			h, err := driver{}.Open(t.Context(), "", "default", cfg, nil, nil, zap.NewNop())
			require.NoError(t, err)
			ai, _ := h.AsAI("default")
			_, err = ai.Complete(t.Context(), &drivers.CompleteOptions{
				Messages: []*aiv1.CompletionMessage{{Role: "user", Content: []*aiv1.ContentBlock{text("hi")}}},
			})
			require.NoError(t, err)

			var body map[string]any
			require.NoError(t, json.Unmarshal(fake.body, &body))
			if tc.want == nil {
				require.NotContains(t, body, "inferenceConfig")
			} else {
				require.Equal(t, tc.want, body["inferenceConfig"])
			}
		})
	}
}

// fakeBedrock is a fake Bedrock Runtime server that records the last request.
type fakeBedrock struct {
	auth string
	body []byte
}

// newFakeBedrock starts a fake Bedrock Runtime server and points the AWS SDK at it.
// It also isolates the test from the host's AWS config.
// The server returns a response with a text block and a tool call.
func newFakeBedrock(t *testing.T) *fakeBedrock {
	fake := &fakeBedrock{}
	// Use TLS since the SDK refuses to send bearer tokens over plain HTTP.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.auth = r.Header.Get("Authorization")
		fake.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output": {"message": {"role": "assistant", "content": [
				{"text": "Checking."},
				{"toolUse": {"toolUseId": "call_1", "name": "query", "input": {"sql": "SELECT 1", "limit": 5}}}
			]}},
			"stopReason": "tool_use",
			"usage": {"inputTokens": 10, "outputTokens": 3, "totalTokens": 13, "cacheReadInputTokens": 4}
		}`))
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	err := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw}), 0o600)
	require.NoError(t, err)
	t.Setenv("AWS_CA_BUNDLE", caPath)
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(dir, "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "")
	t.Setenv("AWS_ENDPOINT_URL_BEDROCK_RUNTIME", srv.URL)
	return fake
}
