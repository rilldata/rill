package jwt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Issuer
type Issuer struct {
	issuerURL string
	jwks      *keyfunc.JWKS
}

// NewIssuer
func NewIssuer(issuerURL, jwksJSON string) (*Issuer, error) {
	jwks, err := keyfunc.NewJSON([]byte(jwksJSON))
	if err != nil {
		return nil, err
	}

	// TODO

	return &Issuer{
		issuerURL: issuerURL,
		jwks:      jwks,
	}, nil
}

func (i *Issuer) NewToken(audienceURL string, systemPerms []Permission, instancePerms map[string][]Permission) string {
	// TODO
	return ""
}

func (i *Issuer) NewSystemToken(audienceURL string, perms []Permission) string {
	return i.NewToken(audienceURL, perms, nil)
}

func (i *Issuer) NewInstanceToken(audienceURL, instanceID string, perms []Permission) string {
	return i.NewToken(audienceURL, nil, map[string][]Permission{instanceID: perms})
}

// WellKnownHandler should be served on {issuerURL}/.well-known/jwks.json
func (i *Issuer) WellKnownHandleFunc(writer http.ResponseWriter, request *http.Request) {
	// TODO
}

// Audience represents a receiver of tokens from Issuer.
// It parses and validates tokens and resolves permissions.
// It refreshes its JWKS in the background from {issuerURL}/.well-known/jwks.json.
type Audience struct {
	issuerURL   string
	audienceURL string
	jwks        *keyfunc.JWKS
}

// OpenAudience creates an Audience. Remember to call Close() when done.
// The issuerURL should be the external URL of the issuing admin server.
// The issuerURL is expected to serve a JWKS on /.well-known/jwks.json.
// The audienceURL should be the external URL of the receiving runtime server.
func OpenAudience(logger *zap.Logger, issuerURL, audienceURL string) (*Audience, error) {
	// To be safe, require issuer and audience is provided
	if issuerURL == "" {
		return nil, fmt.Errorf("issuerURL is not set")
	}
	if audienceURL == "" {
		return nil, fmt.Errorf("audienceURL is not set")
	}

	// The JWKS is assumed to be served on issuer
	jwksURL, err := url.JoinPath(issuerURL, ".well-known/jwks.json")
	if err != nil {
		return nil, err
	}

	// Setup keyfunc that refreshes the JWKS in the background
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		Ctx: context.Background(),
		RefreshErrorHandler: func(err error) {
			logger.Error("JWK refresh failed", zap.Error(err))
		},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, err
	}

	return &Audience{
		issuerURL: issuerURL,
		jwks:      jwks,
	}, nil
}

// Close stops background refresh of the JWKS.
func (a *Audience) Close() {
	a.jwks.EndBackground()
}

// ParseAndValidate parses and validates a JWT and returns Claims if successful.
func (a *Audience) ParseAndValidate(tokenStr string) (Claims, error) {
	claims := &jwtClaims{}

	_, err := jwt.ParseWithClaims(tokenStr, claims, a.jwks.Keyfunc) // NOTE: also validates claims
	if err != nil {
		return nil, err
	}

	if !claims.VerifyIssuer(a.issuerURL, true) {
		return nil, fmt.Errorf("invalid token issuer %q (expected %q)", claims.Issuer, a.issuerURL)
	}

	if !claims.VerifyAudience(a.audienceURL, true) {
		return nil, fmt.Errorf("invalid token audience %q (expected %q)", claims.Audience, a.audienceURL)
	}

	return claims, nil
}
