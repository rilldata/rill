package salesforce

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	force "github.com/ForceCLI/force/lib"
)

const defaultEndpoint = "https://login.salesforce.com"

type authenticationOptions struct {
	Endpoint     string
	Username     string
	Password     string
	JWT          string
	ConnectedApp string
	ClientSecret string
}

// authMode describes which OAuth flow authenticate() will use for a given set
// of options. It is exported via selectAuthMode for unit testing.
type authMode int

const (
	authModeUnknown authMode = iota
	authModeJWT
	authModePassword
	authModeClientCredentials
)

func authenticate(options authenticationOptions) (*force.Force, error) {
	if options.ConnectedApp == "" {
		return nil, fmt.Errorf("connected app client id is required")
	}
	force.ClientId = options.ConnectedApp

	endpoint, err := endpoint(options)
	if err != nil {
		return nil, err
	}

	switch selectAuthMode(options) {
	case authModeJWT:
		if options.Username == "" {
			return nil, fmt.Errorf("username is required for JWT authentication")
		}
		return jwtLogin(endpoint, options)
	case authModePassword:
		if options.ClientSecret == "" {
			return nil, fmt.Errorf("client_secret is required for username/password authentication")
		}
		return passwordFlowLogin(endpoint, options)
	case authModeClientCredentials:
		return clientCredentialsLogin(endpoint, options)
	}
	return nil, fmt.Errorf("unable to authenticate: provide a JWT key, a username and password (with client_secret), or a client_secret for the client credentials flow")
}

// selectAuthMode picks an OAuth flow based on which credentials are populated.
// JWT wins when a key is present; otherwise username+password selects the
// password flow; otherwise a client_secret selects client credentials.
func selectAuthMode(options authenticationOptions) authMode {
	switch {
	case options.JWT != "":
		return authModeJWT
	case options.Username != "" && options.Password != "":
		return authModePassword
	case options.ClientSecret != "":
		return authModeClientCredentials
	}
	return authModeUnknown
}

func endpoint(options authenticationOptions) (endpoint string, err error) {
	isEndpointSelected := options.Endpoint != ""

	if !isEndpointSelected {
		return defaultEndpoint, nil
	}

	// URL needs to have scheme lest the force cli lib chokes
	uri, err := url.Parse(options.Endpoint)
	if err != nil {
		return defaultEndpoint, errors.New("unable to parse endpoint: " + options.Endpoint)
	}

	if uri.Scheme == "" {
		uri.Scheme = "https"
	}

	return uri.String(), nil
}

func jwtLogin(endpoint string, options authenticationOptions) (*force.Force, error) {
	key, err := decodeJWTKey(options.JWT)
	if err != nil {
		return nil, err
	}

	tempfile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("creating tempfile to write rsa key failed: %w", err)
	}
	defer os.Remove(tempfile.Name())

	if _, err = tempfile.Write(key); err != nil {
		return nil, fmt.Errorf("writing rsa key to tempfile failed: %w", err)
	}

	assertion, err := force.JwtAssertionForEndpoint(endpoint, options.Username, tempfile.Name(), options.ConnectedApp)
	if err != nil {
		return nil, err
	}
	session, err := force.JWTLoginAtEndpoint(endpoint, assertion)
	if err != nil {
		return nil, fmt.Errorf("JWT authentication failed: %w", err)
	}

	return force.NewForce(&session), nil
}

func passwordFlowLogin(endpoint string, options authenticationOptions) (*force.Force, error) {
	session, err := force.PasswordFlowLoginAtEndpoint(endpoint, options.ConnectedApp, options.ClientSecret, options.Username, options.Password)
	if err != nil {
		return nil, fmt.Errorf("OAuth password authentication failed: %w", err)
	}
	return force.NewForce(&session), nil
}

func clientCredentialsLogin(endpoint string, options authenticationOptions) (*force.Force, error) {
	session, err := force.ClientCredentialsLoginAtEndpoint(endpoint, options.ConnectedApp, options.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("client credentials authentication failed: %w", err)
	}
	return force.NewForce(&session), nil
}

// decodeJWTKey returns the raw PEM bytes for the JWT private key. The UI
// base64-encodes uploaded keys before writing them to .env so embedded
// newlines don't break the dotenv parser; raw PEM is accepted for
// backwards compatibility with hand-written configs.
func decodeJWTKey(key string) ([]byte, error) {
	trimmed := strings.TrimSpace(key)
	if strings.HasPrefix(trimmed, "-----BEGIN") {
		return []byte(key), nil
	}
	// Tolerate whitespace introduced by line wrapping in .env or YAML.
	compact := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n':
			return -1
		}
		return r
	}, trimmed)
	decoded, err := base64.StdEncoding.DecodeString(compact)
	if err != nil {
		return nil, fmt.Errorf("JWT private key is neither PEM nor valid base64: %w", err)
	}
	return decoded, nil
}
