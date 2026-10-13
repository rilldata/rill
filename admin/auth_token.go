package admin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"golang.org/x/crypto/bcrypt"
)

// Cache settings for s.authCache.
const (
	_authCacheSize = 500              // Number of tokens to cache
	_authCacheTTL  = 10 * time.Second // How long to cache a token before revalidating
)

// AuthToken is the interface package admin uses to provide a consolidated view of a token string and its DB model.
type AuthToken interface {
	Token() *authtoken.Token
	TokenModel() any
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

func (t *userAuthToken) TokenModel() any {
	return t.model
}

func (t *userAuthToken) OwnerID() string {
	if t.model.RepresentingUserID != nil {
		return *t.model.RepresentingUserID
	}

	return t.model.UserID
}

// IssueUserAuthToken generates and persists a new auth token for a user.
func (s *Service) IssueUserAuthToken(ctx context.Context, userID, clientID, displayName string, representingUserID *string, ttl *time.Duration, refresh bool) (AuthToken, error) {
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
		Refresh:            refresh,
		ExpiresOn:          expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &userAuthToken{model: uat, token: tkn}, nil
}

// serviceAuthToken implements AuthToken for tokens belonging to a service.
type serviceAuthToken struct {
	model *database.ServiceAuthToken
	token *authtoken.Token
}

func (t *serviceAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *serviceAuthToken) TokenModel() any {
	return t.model
}

func (t *serviceAuthToken) OwnerID() string {
	return t.model.ServiceID
}

// IssueServiceAuthToken generates and persists a new auth token for a service.
func (s *Service) IssueServiceAuthToken(ctx context.Context, serviceID string, ttl *time.Duration) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeService)

	var expiresOn *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresOn = &t
	}

	sat, err := s.DB.InsertServiceAuthToken(ctx, &database.InsertServiceAuthTokenOptions{
		ID:         tkn.ID.String(),
		SecretHash: tkn.SecretHash(),
		ServiceID:  serviceID,
		ExpiresOn:  expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &serviceAuthToken{model: sat, token: tkn}, nil
}

// deploymentAuthToken implements AuthToken for tokens belonging to a deployment.
type deploymentAuthToken struct {
	model *database.DeploymentAuthToken
	token *authtoken.Token
}

func (t *deploymentAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *deploymentAuthToken) TokenModel() any {
	return t.model
}

func (t *deploymentAuthToken) OwnerID() string {
	return t.model.DeploymentID
}

// IssueDeploymentAuthToken generates and persists a new auth token for a deployment.
func (s *Service) IssueDeploymentAuthToken(ctx context.Context, deploymentID string, ttl *time.Duration) (AuthToken, error) {
	tkn := authtoken.NewRandom(authtoken.TypeDeployment)

	var expiresOn *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresOn = &t
	}

	dat, err := s.DB.InsertDeploymentAuthToken(ctx, &database.InsertDeploymentAuthTokenOptions{
		ID:           tkn.ID.String(),
		SecretHash:   tkn.SecretHash(),
		DeploymentID: deploymentID,
		ExpiresOn:    expiresOn,
	})
	if err != nil {
		return nil, err
	}

	return &deploymentAuthToken{model: dat, token: tkn}, nil
}

// magicAuthToken implements AuthToken for magic tokens belonging to a project.
// It also represents session tokens of password-protected magic tokens (see UnlockMagicAuthToken):
// in that case, token is the session token and model is the magic token it belongs to.
type magicAuthToken struct {
	model *database.MagicAuthToken
	token *authtoken.Token
}

func (t *magicAuthToken) Token() *authtoken.Token {
	return t.token
}

func (t *magicAuthToken) TokenModel() any {
	return t.model
}

func (t *magicAuthToken) OwnerID() string {
	return t.model.ID
}

// IssueMagicAuthTokenOptions provides options for IssueMagicAuthToken.
type IssueMagicAuthTokenOptions struct {
	ProjectID              string
	TTL                    *time.Duration
	CreatedByUserID        *string
	Attributes             map[string]any
	MetricsViewFilterJSONs map[string]string
	Fields                 []string
	State                  string
	DisplayName            string
	// Password optionally requires recipients to enter a password before the token grants access.
	// A password-protected token must be unlocked with UnlockMagicAuthToken before it can be used.
	Password  string
	Internal  bool
	Resources []database.ResourceName
}

// IssueMagicAuthToken generates and persists a new magic auth token for a project.
func (s *Service) IssueMagicAuthToken(ctx context.Context, opts *IssueMagicAuthTokenOptions) (AuthToken, error) {
	for mv := range opts.MetricsViewFilterJSONs {
		if mv == "" {
			return nil, fmt.Errorf("metrics view name cannot be empty in MetricsViewFilterJSONs")
		}
	}

	tkn := authtoken.NewRandom(authtoken.TypeMagic)

	var expiresOn *time.Time
	if opts.TTL != nil {
		t := time.Now().Add(*opts.TTL)
		expiresOn = &t
	}

	var passwordHash []byte
	if opts.Password != "" {
		var err error
		passwordHash, err = HashMagicAuthTokenPassword(opts.Password)
		if err != nil {
			return nil, err
		}
	}

	dat, err := s.DB.InsertMagicAuthToken(ctx, &database.InsertMagicAuthTokenOptions{
		ID:                     tkn.ID.String(),
		SecretHash:             tkn.SecretHash(),
		Secret:                 tkn.Secret[:],
		PasswordHash:           passwordHash,
		ProjectID:              opts.ProjectID,
		ExpiresOn:              expiresOn,
		CreatedByUserID:        opts.CreatedByUserID,
		Attributes:             opts.Attributes,
		MetricsViewFilterJSONs: opts.MetricsViewFilterJSONs,
		Fields:                 opts.Fields,
		State:                  opts.State,
		DisplayName:            opts.DisplayName,
		Internal:               opts.Internal,
		Resources:              opts.Resources,
	})
	if err != nil {
		return nil, err
	}

	return &magicAuthToken{model: dat, token: tkn}, nil
}

// Length constraints for passwords on magic auth tokens. The upper bound is bcrypt's input limit in bytes.
const (
	MagicAuthTokenPasswordMinLength = 8
	MagicAuthTokenPasswordMaxLength = 72
)

// HashMagicAuthTokenPassword hashes a password for storage in a magic auth token.
func HashMagicAuthTokenPassword(password string) ([]byte, error) {
	if len(password) > MagicAuthTokenPasswordMaxLength {
		return nil, fmt.Errorf("password exceeds %d bytes", MagicAuthTokenPasswordMaxLength)
	}
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// Errors returned by UnlockMagicAuthToken.
var (
	ErrMagicAuthTokenNotPasswordProtected = errors.New("magic auth token is not password protected")
	ErrMagicAuthTokenPasswordIncorrect    = errors.New("incorrect magic auth token password")
)

// UnlockMagicAuthToken verifies the password of a password-protected magic auth token and issues a session token for it.
// The session grants the same access as the magic auth token (see magicAuthToken);
// it expires after ttl, or when the magic auth token expires, whichever comes first.
//
// The password is checked against the hash currently in the database, never against a token model from the auth cache,
// and the check and the insert of the session run in one transaction that locks the magic auth token's row.
// A concurrent password change, which also deletes the token's sessions, therefore either commits before the check,
// so that the new password is required, or waits until the session has been inserted, so that it deletes the session too.
// Either way, a session can't be obtained with an old password and outlive it.
func (s *Service) UnlockMagicAuthToken(ctx context.Context, id, password string, ttl time.Duration) (*database.MagicAuthTokenSession, *authtoken.Token, error) {
	ctx, tx, err := s.DB.NewTx(ctx, false)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	mat, err := s.DB.FindMagicAuthTokenForUpdate(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if len(mat.PasswordHash) == 0 {
		return nil, nil, ErrMagicAuthTokenNotPasswordProtected
	}
	if bcrypt.CompareHashAndPassword(mat.PasswordHash, []byte(password)) != nil {
		return nil, nil, ErrMagicAuthTokenPasswordIncorrect
	}

	tkn := authtoken.NewRandom(authtoken.TypeMagicSession)

	expiresOn := time.Now().Add(ttl)
	if mat.ExpiresOn != nil && mat.ExpiresOn.Before(expiresOn) {
		expiresOn = *mat.ExpiresOn
	}

	sess, err := s.DB.InsertMagicAuthTokenSession(ctx, &database.InsertMagicAuthTokenSessionOptions{
		ID:               tkn.ID.String(),
		SecretHash:       tkn.SecretHash(),
		MagicAuthTokenID: mat.ID,
		ExpiresOn:        expiresOn,
	})
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	return sess, tkn, nil
}

// ExtendBrowserSessionAuthToken extends a Rill web browser session token when its
// remaining lifetime is at or below refreshThreshold.
func (s *Service) ExtendBrowserSessionAuthToken(ctx context.Context, authTok AuthToken, fullTTL, refreshThreshold time.Duration) error {
	uat, ok := authTok.TokenModel().(*database.UserAuthToken)
	if !ok {
		return nil
	}
	if uat.AuthClientID == nil || *uat.AuthClientID != database.AuthClientIDRillWeb {
		return nil
	}
	if uat.RepresentingUserID != nil || uat.Refresh {
		return nil
	}
	if uat.ExpiresOn != nil && time.Until(*uat.ExpiresOn) > refreshThreshold {
		return nil
	}
	newExpiresOn := time.Now().Add(fullTTL)
	return s.DB.UpdateUserAuthTokenExpiresOn(ctx, uat.ID, newExpiresOn)
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
	case authtoken.TypeService:
		return s.DB.DeleteServiceAuthToken(ctx, parsed.ID.String())
	case authtoken.TypeDeployment:
		return fmt.Errorf("deployment auth tokens cannot be revoked")
	case authtoken.TypeMagic:
		return s.DB.DeleteMagicAuthToken(ctx, parsed.ID.String())
	case authtoken.TypeMagicSession:
		return s.DB.DeleteMagicAuthTokenSession(ctx, parsed.ID.String())
	default:
		return fmt.Errorf("unknown auth token type %q", parsed.Type)
	}
}

// PurgeAuthTokenCache purges the short-term in-memory auth token cache.
func (s *Service) PurgeAuthTokenCache() {
	s.authCache.Purge()
}

// ValidateAuthToken validates an auth token against persistent storage.
// It includes a short-term in-memory cache to prevent
func (s *Service) ValidateAuthToken(ctx context.Context, token string) (AuthToken, error) {
	// Wrapper type for cache entries
	type authCacheEntry struct {
		token AuthToken
		err   error
		time  time.Time
	}

	// Use a secure hash of the token string as the cache key to avoid storing raw tokens in memory.
	tokenHash := sha256.Sum256([]byte(token))
	cacheKey := fmt.Sprintf("%x", tokenHash[:])

	// Try cache
	val, ok := s.authCache.Get(cacheKey)
	if ok {
		entry, ok := val.(authCacheEntry)
		if ok && time.Since(entry.time) < _authCacheTTL {
			return entry.token, entry.err
		}
		// Even if its expired, we don't remove it from the cache as it'll be replaced below.
	}

	// Not cached or expired, validate
	authTok, err := s.validateAuthTokenUncached(ctx, token)
	if err != nil && errors.Is(err, ctx.Err()) { // Only exit early for context errors
		return nil, err
	}

	// Cache the validation result (both token and error)
	s.authCache.Add(cacheKey, authCacheEntry{
		token: authTok,
		err:   err,
		time:  time.Now(),
	})

	return authTok, err
}

func (s *Service) validateAuthTokenUncached(ctx context.Context, token string) (AuthToken, error) {
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

		s.Used.UserToken(uat.ID)
		s.Used.User(uat.UserID)
		if uat.AuthClientID != nil {
			s.Used.Client(*uat.AuthClientID)
		}

		return &userAuthToken{model: uat, token: parsed}, nil
	case authtoken.TypeService:
		sat, err := s.DB.FindServiceAuthToken(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if sat.ExpiresOn != nil && sat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(sat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.ServiceToken(sat.ID)
		s.Used.Service(sat.ServiceID)

		return &serviceAuthToken{model: sat, token: parsed}, nil
	case authtoken.TypeDeployment:
		dat, err := s.DB.FindDeploymentAuthToken(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if dat.ExpiresOn != nil && dat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(dat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.DeploymentToken(dat.ID)

		return &deploymentAuthToken{model: dat, token: parsed}, nil
	case authtoken.TypeMagic:
		mat, err := s.DB.FindMagicAuthToken(ctx, parsed.ID.String(), false)
		if err != nil {
			return nil, err
		}

		if mat.ExpiresOn != nil && mat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(mat.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		s.Used.MagicAuthToken(mat.ID)

		return &magicAuthToken{model: mat, token: parsed}, nil
	case authtoken.TypeMagicSession:
		sess, err := s.DB.FindMagicAuthTokenSession(ctx, parsed.ID.String())
		if err != nil {
			return nil, err
		}

		if sess.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		if !bytes.Equal(sess.SecretHash, parsed.SecretHash()) {
			return nil, fmt.Errorf("invalid auth token")
		}

		// The session acts as the magic token it belongs to.
		mat, err := s.DB.FindMagicAuthToken(ctx, sess.MagicAuthTokenID, false)
		if err != nil {
			return nil, err
		}

		if mat.ExpiresOn != nil && mat.ExpiresOn.Before(time.Now()) {
			return nil, fmt.Errorf("auth token is expired")
		}

		s.Used.MagicAuthToken(mat.ID)

		return &magicAuthToken{model: mat, token: parsed}, nil
	default:
		return nil, fmt.Errorf("unknown auth token type %q", parsed.Type)
	}
}
