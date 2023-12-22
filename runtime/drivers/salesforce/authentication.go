package salesforce

import (
	"errors"
	"fmt"
	"net/url"

	force "github.com/ForceCLI/force/lib"
)

const defaultEndpoint = "https://login.salesforce.com"

type authenticationOptions struct {
	Endpoint     string
	Username     string
	Password     string
	JWT          string
	ConnectedApp string
}

// Authenticable represents the authentation subset of the force api
type Authenticable interface {
	JwtAssertionForEndpoint(endpoint string, username string, keyfile string, clientID string) (string, error)
	JwtLoginAtEndpoint(endpoint string, assertion string) (*force.Force, error)
	ForceSoapLoginAtEndpoint(endpoint string, userName string, password string) (*force.Force, error)
}

// a concrete implementation of Authenticable
type forceProvider struct{}

func authenticate(options authenticationOptions) (*force.Force, error) {
	if options.ConnectedApp == "" {
		options.ConnectedApp = defaultClientID
	}
	force.ClientId = options.ConnectedApp
	provider := forceProvider{}

	isUsernameSelected := len(options.Username) > 0
	isJWTSelected := len(options.JWT) > 0
	isSOAPSelected := len(options.Password) > 0

	endpoint, err := endpoint(options)
	if err != nil {
		return nil, err
	}

	switch {
	case isJWTSelected:
		return jwtLogin(endpoint, options, provider)
	case isSOAPSelected:
		return provider.ForceSoapLoginAtEndpoint(endpoint, options.Username, options.Password)
	case !isUsernameSelected:
		return nil, fmt.Errorf("username missing")
	}
	return nil, fmt.Errorf("unable to authenticate")
}

func endpoint(options authenticationOptions) (endpoint string, err error) {
	isEndpointSelected := len(options.Endpoint) > 0

	if isEndpointSelected {
		// URL needs to have scheme lest the force cli lib chokes
		uri, err := url.Parse(options.Endpoint)
		if err != nil {
			return defaultEndpoint, errors.New("Unable to parse endpoint: " + options.Endpoint)
		}

		if uri.Host == "" {
			uri, err = url.Parse("https://" + options.Endpoint)
			if err != nil {
				return defaultEndpoint, errors.New("Could not identify host: " + options.Endpoint)
			}
		}

		return uri.String(), nil
	}

	return defaultEndpoint, nil
}

func jwtLogin(endpoint string, options authenticationOptions, provider Authenticable) (*force.Force, error) {
	assertion, err := provider.JwtAssertionForEndpoint(endpoint, options.Username, options.JWT, options.ConnectedApp)
	if err != nil {
		return nil, err
	}
	return provider.JwtLoginAtEndpoint(endpoint, assertion)
}

func (provider forceProvider) JwtAssertionForEndpoint(endpoint, username, keyfile, clientID string) (string, error) {
	return force.JwtAssertionForEndpoint(endpoint, username, keyfile, clientID)
}

func (provider forceProvider) JwtLoginAtEndpoint(endpoint, assertion string) (*force.Force, error) {
	session, err := force.JWTLoginAtEndpoint(endpoint, assertion)
	if err != nil {
		return nil, fmt.Errorf("JWT authentication failed: %w", err)
	}

	return force.NewForce(&session), nil
}

func (provider forceProvider) ForceSoapLoginAtEndpoint(endpoint, username, password string) (*force.Force, error) {
	session, err := force.ForceSoapLoginAtEndpoint(endpoint, username, password)
	if err != nil {
		return nil, fmt.Errorf("SOAP authentication failed: %w", err)
	}

	return force.NewForce(&session), nil
}
