package salesforce

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

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

func authenticate(options authenticationOptions) (*force.Force, error) {
	if options.ConnectedApp == "" {
		return nil, fmt.Errorf("connected app client id is required")
	}
	force.ClientId = options.ConnectedApp

	isUsernameSelected := len(options.Username) > 0
	isJWTSelected := len(options.JWT) > 0
	isSOAPSelected := len(options.Password) > 0

	endpoint, err := endpoint(options)
	if err != nil {
		return nil, err
	}

	switch {
	case isJWTSelected:
		return jwtLogin(endpoint, options)
	case isSOAPSelected:
		return soapLoginAtEndpoint(endpoint, options.Username, options.Password)
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
			return defaultEndpoint, errors.New("unable to parse endpoint: " + options.Endpoint)
		}

		if uri.Host == "" {
			uri, err = url.Parse("https://" + options.Endpoint)
			if err != nil {
				return defaultEndpoint, errors.New("could not identify host: " + options.Endpoint)
			}
		}

		return uri.String(), nil
	}

	return defaultEndpoint, nil
}

func jwtLogin(endpoint string, options authenticationOptions) (*force.Force, error) {
	tempfile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, fmt.Errorf("creating tempfile to write rsa key failed: %w", err)
	}
	defer os.Remove(tempfile.Name())

	if _, err = tempfile.WriteString(options.JWT); err != nil {
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

func soapLoginAtEndpoint(endpoint, username, password string) (*force.Force, error) {
	session, err := force.ForceSoapLoginAtEndpoint(endpoint, username, password)
	if err != nil {
		return nil, fmt.Errorf("SOAP authentication failed: %w", err)
	}

	return force.NewForce(&session), nil
}
