package git

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

func NewInstallationToken(installationID int64, appKey []byte, githubAppID string) (*github.InstallationToken, error) {
	tokenSrc, err := newTokenSource(appKey, githubAppID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, tokenSrc))
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	return installationToken, err
}

type tokenSource struct {
	key         *rsa.PrivateKey
	githubAppID string
}

func newTokenSource(pemBytes []byte, githubAppID string) (oauth2.TokenSource, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid pem key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &tokenSource{key: key, githubAppID: githubAppID}, nil
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	// issued at time, 60 seconds in the past to allow for clock drift
	iat := time.Now().Add(0 - time.Minute)
	// JWT expiration time (10 minute maximum)
	expiry := iat.Add(time.Minute * 10)
	// Create the Claims
	claims := &jwt.RegisteredClaims{
		IssuedAt: jwt.NewNumericDate(iat),
		// JWT expiration time (10 minute maximum)
		ExpiresAt: jwt.NewNumericDate(expiry),
		// GitHub App's identifier
		Issuer: t.githubAppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(t.key)
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{AccessToken: ss, Expiry: expiry}, nil
}
