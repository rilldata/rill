package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConnectorOAuthCaller(t *testing.T) {
	tests := []struct {
		name      string
		claims    *runtime.SecurityClaims
		adminTok  string
		adminErr  error
		wantTok   *drivers.OAuthToken
		wantErr   error
		wantAuthz bool
	}{
		{
			name:     "interactive user",
			claims:   &runtime.SecurityClaims{UserID: "u", InteractiveUser: true},
			adminTok: "t",
			wantTok:  &drivers.OAuthToken{UserID: "u", AccessToken: "t", ExpiresOn: fakeOAuthExpiresOn},
		},
		{
			name:      "interactive user without authorization",
			claims:    &runtime.SecurityClaims{UserID: "u", InteractiveUser: true},
			adminErr:  drivers.ErrNotFound,
			wantAuthz: true,
		},
		{
			name:     "view as",
			claims:   &runtime.SecurityClaims{UserID: "mocked", InteractiveUser: false},
			adminTok: "t",
			wantErr:  drivers.ErrOAuthUserRequired,
		},
		{
			name:     "service",
			claims:   &runtime.SecurityClaims{UserID: "svc", SkipChecks: true},
			adminTok: "t",
			wantErr:  drivers.ErrOAuthUserRequired,
		},
		{
			name:     "rill developer",
			claims:   &runtime.SecurityClaims{},
			adminErr: drivers.ErrNotImplemented,
			wantErr:  drivers.ErrNotImplemented,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			admin := &fakeOAuthAdmin{tok: tt.adminTok, err: tt.adminErr}
			c := connectorOAuthCaller{
				admin: func(ctx context.Context, instanceID string) (drivers.AdminService, func(), error) {
					return admin, func() {}, nil
				},
			}
			ctx := auth.WithClaims(context.Background(), tt.claims)

			tok, err := c.Token(ctx, "inst", "dbx", "hash", false)
			switch {
			case tt.wantAuthz:
				var authzErr *drivers.OAuthAuthorizationRequiredError
				require.ErrorAs(t, err, &authzErr)
				require.Equal(t, "dbx", authzErr.Connector)
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				require.Equal(t, tt.wantTok, tok)
			}

			if tt.claims.UserID != "" && !tt.claims.InteractiveUser {
				require.NotContains(t, admin.userIDs, tt.claims.UserID)
			}
		})
	}
}

func TestMapGRPCErrorOAuth(t *testing.T) {
	err := mapGRPCError(fmt.Errorf("query failed: %w", &drivers.OAuthAuthorizationRequiredError{Connector: "dbx"}))
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	require.Equal(t, `sign in to connector "dbx" to view this data`, st.Message())

	err = mapGRPCError(fmt.Errorf("query failed: %w", drivers.ErrOAuthUserRequired))
	st, ok = status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	require.Equal(t, drivers.ErrOAuthUserRequired.Error(), st.Message())
}

var fakeOAuthExpiresOn = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

// fakeOAuthAdmin is a drivers.AdminService that serves a fixed connector OAuth token and records the users it was requested for.
type fakeOAuthAdmin struct {
	drivers.AdminService
	tok     string
	err     error
	userIDs []string
}

func (a *fakeOAuthAdmin) GetConnectorOAuthToken(ctx context.Context, connector, configHash, userID string, forceRefresh bool) (*drivers.MCPOAuthToken, error) {
	a.userIDs = append(a.userIDs, userID)
	if a.err != nil {
		return nil, a.err
	}
	if connector != "dbx" || configHash != "hash" {
		return nil, errors.New("unexpected connector")
	}
	return &drivers.MCPOAuthToken{AccessToken: a.tok, ExpiresOn: fakeOAuthExpiresOn}, nil
}
