package admin

import (
	"testing"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/stretchr/testify/require"
)

func TestEmbedToken(t *testing.T) {
	issuer, err := auth.NewEphemeralIssuer("http://admin.example.com")
	require.NoError(t, err)
	s := &Service{issuer: issuer}

	resources := []database.ResourceName{{Type: runtime.ResourceKindExplore, Name: "e1"}}
	token, err := s.IssueEmbedToken("p1", "ext_abc", resources, time.Hour)
	require.NoError(t, err)

	res, err := s.ValidateEmbedToken(token)
	require.NoError(t, err)
	require.Equal(t, "p1", res.ProjectID)
	require.Equal(t, "ext_abc", res.EmbedUserID)
	require.Equal(t, "ext_abc", res.Subject)
	require.Equal(t, resources, res.Resources)

	// Expired tokens are rejected
	token, err = s.IssueEmbedToken("p1", "ext_abc", nil, -time.Minute)
	require.NoError(t, err)
	_, err = s.ValidateEmbedToken(token)
	require.Error(t, err)

	// Runtime tokens signed by the same issuer are rejected
	token, err = issuer.NewToken(auth.TokenOptions{
		AudienceURL: "http://runtime.example.com",
		Subject:     "ext_abc",
		TTL:         time.Hour,
	})
	require.NoError(t, err)
	_, err = s.ValidateEmbedToken(token)
	require.Error(t, err)
}
