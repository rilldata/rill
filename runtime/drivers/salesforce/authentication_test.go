package salesforce

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpoint(t *testing.T) {
	e, err := endpoint(authenticationOptions{Endpoint: "login.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://login.salesforce.com", e)

	e, err = endpoint(authenticationOptions{Endpoint: "example.my.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://example.my.salesforce.com", e)

	e, err = endpoint(authenticationOptions{Endpoint: "https://login.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://login.salesforce.com", e)

	e, err = endpoint(authenticationOptions{Endpoint: "https://example.my.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://example.my.salesforce.com", e)
}

func TestSelectAuthMode(t *testing.T) {
	cases := []struct {
		name string
		opts authenticationOptions
		want authMode
	}{
		{
			name: "jwt wins when key is present",
			opts: authenticationOptions{
				Username:     "user@example.com",
				Password:     "pw",
				JWT:          "key",
				ConnectedApp: "cid",
				ClientSecret: "secret",
			},
			want: authModeJWT,
		},
		{
			name: "password flow when username and password are set",
			opts: authenticationOptions{
				Username:     "user@example.com",
				Password:     "pw",
				ConnectedApp: "cid",
				ClientSecret: "secret",
			},
			want: authModePassword,
		},
		{
			name: "client credentials when only client_id and secret are set",
			opts: authenticationOptions{
				ConnectedApp: "cid",
				ClientSecret: "secret",
			},
			want: authModeClientCredentials,
		},
		{
			name: "client credentials when username is set without password",
			opts: authenticationOptions{
				Username:     "user@example.com",
				ConnectedApp: "cid",
				ClientSecret: "secret",
			},
			want: authModeClientCredentials,
		},
		{
			name: "unknown when nothing useful is provided",
			opts: authenticationOptions{
				Username:     "user@example.com",
				ConnectedApp: "cid",
			},
			want: authModeUnknown,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, selectAuthMode(c.opts))
		})
	}
}

func TestParseSourceProperties(t *testing.T) {
	// Explicit soql is the historical shape; Bulk API 2.0 derives the
	// SObject from the query itself so `sobject` is no longer required.
	conf, err := parseSourceProperties(map[string]any{
		"soql": "SELECT Id FROM Opportunity",
	})
	require.NoError(t, err)
	require.Equal(t, "SELECT Id FROM Opportunity", conf.SOQL)

	// `sql:` is accepted as a fallback for `soql:` so Salesforce fits the
	// warehouse model shape produced by the connector explorer.
	conf, err = parseSourceProperties(map[string]any{
		"sql": "SELECT Id FROM Account",
	})
	require.NoError(t, err)
	require.Equal(t, "SELECT Id FROM Account", conf.SOQL)

	// An explicit soql wins over sql when both are supplied.
	conf, err = parseSourceProperties(map[string]any{
		"soql": "SELECT Id FROM Opportunity",
		"sql":  "SELECT Name FROM Lead",
	})
	require.NoError(t, err)
	require.Equal(t, "SELECT Id FROM Opportunity", conf.SOQL)

	// Missing query is still an error.
	_, err = parseSourceProperties(map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "soql")

	// A legacy `sobject:` value is silently ignored (no schema field) so
	// existing project YAMLs don't error out after the Bulk 2.0 switch.
	conf, err = parseSourceProperties(map[string]any{
		"soql":    "SELECT Id FROM Opportunity",
		"sobject": "Opportunity",
	})
	require.NoError(t, err)
	require.Equal(t, "SELECT Id FROM Opportunity", conf.SOQL)
}

func TestDecodeJWTKey(t *testing.T) {
	pem := "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----\n"

	// Raw PEM is passed through unchanged (backwards compat with hand-written configs).
	out, err := decodeJWTKey(pem)
	require.NoError(t, err)
	require.Equal(t, pem, string(out))

	// Base64-encoded PEM (the shape produced by the UI's file upload) decodes to PEM.
	encoded := base64.StdEncoding.EncodeToString([]byte(pem))
	out, err = decodeJWTKey(encoded)
	require.NoError(t, err)
	require.Equal(t, pem, string(out))

	// Whitespace inside the base64 value is tolerated (e.g. wrapped lines in .env).
	wrapped := encoded[:20] + "\n" + encoded[20:40] + " " + encoded[40:]
	out, err = decodeJWTKey(wrapped)
	require.NoError(t, err)
	require.Equal(t, pem, string(out))

	// Garbage that is neither PEM nor base64 is rejected.
	_, err = decodeJWTKey("not a key !@#$")
	require.Error(t, err)
}

func TestApplyOverrides(t *testing.T) {
	base := authenticationOptions{
		Endpoint:     "login.salesforce.com",
		ConnectedApp: defaultClientID,
		Username:     "config@example.com",
		Password:     "config-pw",
	}

	// Per-source values supplied on the model win over connector config.
	src := &sourceProperties{
		Endpoint: "test.salesforce.com",
		Username: "src@example.com",
		Password: "src-pw",
	}
	got := base
	src.applyOverrides(&got)
	require.Equal(t, "test.salesforce.com", got.Endpoint)
	require.Equal(t, "src@example.com", got.Username)
	require.Equal(t, "src-pw", got.Password)
	// ConnectedApp was not set on the source, so the connector default stays.
	require.Equal(t, defaultClientID, got.ConnectedApp)

	// An empty source struct leaves the connector-level options untouched.
	got = base
	(&sourceProperties{}).applyOverrides(&got)
	require.Equal(t, base, got)
}

func TestAuthenticateValidation(t *testing.T) {
	// Missing connected app client id is always an error.
	_, err := authenticate(authenticationOptions{Username: "u", Password: "p", ClientSecret: "s"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "connected app client id")

	// Username/password without client_secret should be reported.
	_, err = authenticate(authenticationOptions{ConnectedApp: "cid", Username: "u", Password: "p"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "client_secret")

	// No credentials at all returns the catch-all error.
	_, err = authenticate(authenticationOptions{ConnectedApp: "cid"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to authenticate")
}
