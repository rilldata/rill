package admin

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
)

// AuthToken is the interface package admin uses to provide a consolidated view of a token string and its DB model.
type AuthToken interface {
	Token() *authtoken.Token
	OwnerID() string
}

// userAuthToken implements AuthToken for tokens belonging to a user.
type userAuthToken struct {
	model *database.UserAuthToken
	token *authtoken.Token
}

func (t *userAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *userAuthToken) OwnerID() string {
	if t.model.RepresentingUserID != nil {
		return *t.model.RepresentingUserID
	}

	return t.model.UserID
}

// IssueUserAuthToken generates and persists a new auth token for a user.
func (s *Service) IssueUserAuthToken(ctx context.Context, userID, clientID, displayName string, representingUserID *string, ttl *time.Duration) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeUser)

	var expiresOn *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresOn = &t
	}

	uat, err := s.DB.InsertUserAuthToken(ctx, &database.InsertUserAuthTokenOptions{
		ID:                 tkn.ID.String(),
		SecretHash:         tkn.SecretHash(),
		UserID:             userID,
		AuthClientID:       &clientID,
		DisplayName:        displayName,
		RepresentingUserID: representingUserID,
		ExpiresOn:          expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &userAuthToken{model: uat, token: tkn}, nil
}

// ValidateAuthToken validates an auth token against persistent storage.
func (s *Service) ValidateAuthToken(ctx context.Context, token string) (AuthToken, error) {
	parsed, err := authtoken.FromString(token)
	if err != nil {
		return nil, err
	}

	switch parsed.Type {
	case authtoken.TypeUser:
		uat, err := s.DB.FindUserAuthToken(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if uat.ExpiresOn != nil && uat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(uat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		return &userAuthToken{model: uat, token: parsed}, nil
	default:
		return nil, fmt.Errorf("unknown auth token type %q", parsed.Type)
	}
}

// RevokeAuthToken removes an auth token from persistent storage.
func (s *Service) RevokeAuthToken(ctx context.Context, token string) error {
	parsed, err := authtoken.FromString(token)
	if err != nil {
		return err
	}
	switch parsed.Type {
	case authtoken.TypeUser:
		return s.DB.DeleteUserAuthToken(ctx, parsed.ID.String())
	default:
		return fmt.Errorf("unknown auth token type %q", parsed.Type)
	}
}
