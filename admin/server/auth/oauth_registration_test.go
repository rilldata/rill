package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOAuthDynamicRegistrationPolicy(t *testing.T) {
	// Dynamic registration normalizes safe metadata and rejects scopes, grants, redirects, or bodies outside policy.
	t.Run("persists only approved metadata and unique redirects", func(t *testing.T) {
		db := newOAuthPolicyTestDB()
		a := newOAuthPolicyTestAuthenticator(db)
		req := oauth.ClientRegistrationRequest{
			ClientName: "Local MCP client",
			Scope:      " offline_access  offline_access ",
			GrantTypes: []string{
				" " + authorizationCodeGrantType + " ",
				refreshTokenGrantType,
				authorizationCodeGrantType,
			},
			RedirectURIs: []string{
				" http://127.0.0.1:43123/callback ",
				"http://127.0.0.1:43123/callback",
				"http://[::1]:43123/callback",
				"https://app.example.com/callback",
			},
		}

		w := registerOAuthClient(t, a, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var resp oauth.ClientRegistrationResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		stored, err := db.FindAuthClient(t.Context(), resp.ClientID)
		require.NoError(t, err)
		require.Equal(t, "offline_access", stored.Scope)
		require.Equal(t, []string{authorizationCodeGrantType, refreshTokenGrantType}, stored.GrantTypes)
		require.Equal(t, []string{
			"http://127.0.0.1:43123/callback",
			"http://[::1]:43123/callback",
			"https://app.example.com/callback",
		}, stored.RedirectURIs)
		require.Equal(t, stored.Scope, resp.Scope)
		require.Equal(t, stored.GrantTypes, resp.GrantTypes)
		require.Equal(t, stored.RedirectURIs, resp.RedirectURIs)
	})

	tests := []struct {
		name   string
		modify func(*oauth.ClientRegistrationRequest)
		body   string
	}{
		{
			name: "rejects privileged scope from an untrusted registrant",
			modify: func(req *oauth.ClientRegistrationRequest) {
				req.Scope = "offline_access " + longLivedAccessTokenScope
			},
			body: "requires trusted registration",
		},
		{
			name: "rejects unknown scope",
			modify: func(req *oauth.ClientRegistrationRequest) {
				req.Scope = "openid"
			},
			body: "unsupported scope",
		},
		{
			name: "rejects unknown grant",
			modify: func(req *oauth.ClientRegistrationRequest) {
				req.GrantTypes = append(req.GrantTypes, "client_credentials")
			},
			body: "unsupported grant_type",
		},
		{
			name: "rejects plain HTTP for a remote callback",
			modify: func(req *oauth.ClientRegistrationRequest) {
				req.RedirectURIs = []string{"http://app.example.com/callback"}
			},
			body: "must use https unless it is loopback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newOAuthPolicyTestDB()
			a := newOAuthPolicyTestAuthenticator(db)
			req := oauth.ClientRegistrationRequest{
				Scope:        "offline_access",
				GrantTypes:   []string{authorizationCodeGrantType},
				RedirectURIs: []string{"https://app.example.com/callback"},
			}
			tt.modify(&req)

			w := registerOAuthClient(t, a, req)

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.Contains(t, w.Body.String(), tt.body)
			require.Empty(t, db.clients, "invalid metadata must not be persisted")
		})
	}

	t.Run("rejects an oversized body before decoding or persistence", func(t *testing.T) {
		db := newOAuthPolicyTestDB()
		a := newOAuthPolicyTestAuthenticator(db)
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/register", strings.NewReader(strings.Repeat("x", maxOAuthRegistrationBodyBytes+1)))
		w := httptest.NewRecorder()

		a.handleOAuthRegister(w, req)

		require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		require.Empty(t, db.clients)
	})
}

func TestOAuthAuthorizePolicy(t *testing.T) {
	// Authorization must distinguish anonymous login redirects from a valid, grant-bound browser consent action.
	const (
		callback = "https://app.example.com/callback"
		userID   = "user-123"
	)

	t.Run("anonymous response is one login redirect", func(t *testing.T) {
		db := newOAuthPolicyTestDB()
		a := newOAuthPolicyTestAuthenticator(db)
		r := newOAuthAuthorizeRequest(t, anonClaims{}, "client-1", callback)
		w := httptest.NewRecorder()

		a.handleAuthorizeRequest(w, r)

		require.Equal(t, http.StatusTemporaryRedirect, w.Code)
		require.Contains(t, w.Header().Get("Location"), "/auth/login?redirect=")
		require.NotContains(t, w.Body.String(), "only users can be authorized")
		require.Equal(t, 1, strings.Count(w.Body.String(), "Temporary Redirect"))
		require.Empty(t, db.authorizationCodes)
	})

	t.Run("authenticated browser action consents to a registered ordinary client", func(t *testing.T) {
		db := newOAuthPolicyTestDB()
		a := newOAuthPolicyTestAuthenticator(db)
		client, err := db.InsertAuthClient(t.Context(), "MCP client", "offline_access", []string{authorizationCodeGrantType}, []string{callback})
		require.NoError(t, err)
		claims := &oauthPolicyTestClaims{ownerType: OwnerTypeUser, ownerID: userID}
		r := newOAuthAuthorizeRequest(t, claims, client.ID, callback)
		w := httptest.NewRecorder()

		a.handleAuthorizeRequest(w, r)

		require.Equal(t, http.StatusFound, w.Code)
		redirect, err := url.Parse(w.Header().Get("Location"))
		require.NoError(t, err)
		require.Equal(t, "app.example.com", redirect.Host)
		require.Equal(t, "/callback", redirect.Path)
		require.NotEmpty(t, redirect.Query().Get("code"))
		require.Equal(t, "opaque-state", redirect.Query().Get("state"))
		require.Len(t, db.authorizationCodes, 1)
		code := db.authorizationCodes[0]
		require.Equal(t, userID, code.UserID)
		require.Equal(t, client.ID, code.ClientID)
		require.Equal(t, callback, code.RedirectURI)
		require.Equal(t, "test-challenge", code.CodeChallenge)
		require.Equal(t, "S256", code.CodeChallengeMethod)
	})

	t.Run("client without authorization code grant is denied", func(t *testing.T) {
		db := newOAuthPolicyTestDB()
		a := newOAuthPolicyTestAuthenticator(db)
		client, err := db.InsertAuthClient(t.Context(), "Refresh-only client", "offline_access", []string{refreshTokenGrantType}, []string{callback})
		require.NoError(t, err)
		claims := &oauthPolicyTestClaims{ownerType: OwnerTypeUser, ownerID: userID}
		r := newOAuthAuthorizeRequest(t, claims, client.ID, callback)
		w := httptest.NewRecorder()

		a.handleAuthorizeRequest(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "not permitted to use authorization codes")
		require.Empty(t, db.authorizationCodes)
	})
}

func registerOAuthClient(t *testing.T, a *Authenticator, registration oauth.ClientRegistrationRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(registration)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/auth/oauth/register", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	a.handleOAuthRegister(w, req)
	return w
}

func newOAuthAuthorizeRequest(t *testing.T, claims Claims, clientID, redirectURI string) *http.Request {
	t.Helper()
	query := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"code_challenge":        {"test-challenge"},
		"code_challenge_method": {"S256"},
		"state":                 {"opaque-state"},
	}
	req := httptest.NewRequest(http.MethodGet, "/auth/oauth/authorize?"+query.Encode(), nil)
	return req.WithContext(WithClaims(req.Context(), claims))
}

func newOAuthPolicyTestAuthenticator(db *oauthPolicyTestDB) *Authenticator {
	return &Authenticator{
		logger: zap.NewNop(),
		admin:  &admin.Service{DB: db},
	}
}

type oauthPolicyTestClaims struct {
	Claims
	ownerType OwnerType
	ownerID   string
}

func (c *oauthPolicyTestClaims) OwnerType() OwnerType { return c.ownerType }
func (c *oauthPolicyTestClaims) OwnerID() string      { return c.ownerID }

type oauthPolicyTestDB struct {
	database.DB
	clients            map[string]*database.AuthClient
	authorizationCodes []*database.AuthorizationCode
}

func newOAuthPolicyTestDB() *oauthPolicyTestDB {
	return &oauthPolicyTestDB{clients: make(map[string]*database.AuthClient)}
}

func (db *oauthPolicyTestDB) InsertAuthClient(_ context.Context, displayName, scope string, grantTypes, redirectURIs []string) (*database.AuthClient, error) {
	client := &database.AuthClient{
		ID:           fmt.Sprintf("dynamic-client-%d", len(db.clients)+1),
		DisplayName:  displayName,
		Scope:        scope,
		GrantTypes:   append([]string(nil), grantTypes...),
		RedirectURIs: append([]string(nil), redirectURIs...),
		CreatedOn:    time.Unix(1_700_000_000, 0).UTC(),
	}
	db.clients[client.ID] = client
	return client, nil
}

func (db *oauthPolicyTestDB) FindAuthClient(_ context.Context, id string) (*database.AuthClient, error) {
	client, ok := db.clients[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return client, nil
}

func (db *oauthPolicyTestDB) InsertAuthorizationCode(_ context.Context, code, userID, clientID, redirectURI, codeChallenge, codeChallengeMethod string, expiration time.Time) (*database.AuthorizationCode, error) {
	authCode := &database.AuthorizationCode{
		ID:                  fmt.Sprintf("authorization-code-%d", len(db.authorizationCodes)+1),
		Code:                code,
		UserID:              userID,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Expiration:          expiration,
	}
	db.authorizationCodes = append(db.authorizationCodes, authCode)
	return authCode, nil
}
