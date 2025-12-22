package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/rilldata/rill/admin/pkg/oauth"
)

const (
	// characters allowed in the PKCE code verifier
	charset             = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	codeChallengeMethod = "S256"
)

type Authenticator struct {
	client       *http.Client
	baseAuthURL  string
	redirectURL  string
	codeVerifier string
	clientID     string
	OriginURL    string
}

func NewAuthenticator(baseAuthURL, redirectURL, clientID, origin string) (*Authenticator, error) {
	// Generate a new code verifier
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		client:       http.DefaultClient,
		baseAuthURL:  baseAuthURL,
		redirectURL:  redirectURL,
		codeVerifier: codeVerifier,
		clientID:     clientID,
		OriginURL:    origin,
	}, nil
}

func (a *Authenticator) GetAuthURL(state string) string {
	// Create the code challenge from the code verifier
	codeChallenge := createCodeChallenge(a.codeVerifier)
	// Create the authorization request URL
	// Create a new URL instance from the authURL string
	u, _ := url.Parse(a.baseAuthURL + "/auth/oauth/authorize")

	// Create a new query string from the URL's query
	q := u.Query()

	// Set the client_id query parameter
	q.Set("client_id", a.clientID)
	// Set the redirect_uri query parameter
	q.Set("redirect_uri", a.redirectURL)
	// Set the response_type query parameter
	q.Set("response_type", "code")
	// Set the code_challenge query parameter
	q.Set("code_challenge", codeChallenge)
	// Set the code_challenge_method query parameter
	q.Set("code_challenge_method", codeChallengeMethod)
	// Set the state, will be used later to retrieve this authenticator
	q.Set("state", state)

	// Encode the query string
	u.RawQuery = q.Encode()

	// Return the URL as a string
	return u.String()
}

func (a *Authenticator) ExchangeCodeForToken(code string) (string, error) {
	// Create the token request
	req, err := tokenRequest(a.baseAuthURL, code, a.clientID, a.redirectURL, a.codeVerifier)
	if err != nil {
		return "", err
	}

	// Send the token request
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the response is an error
	if resp.StatusCode != http.StatusOK {
		// read body to get the error message
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, status: %s, body: %s", resp.StatusCode, resp.Status, string(body))
	}

	tokenResponse := &oauth.TokenResponse{}
	// Decode the response into the tokenResponse struct
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	// Return the access token
	return tokenResponse.AccessToken, nil
}

func tokenRequest(baseAuthURL, code, clientID, redirectURI, codeVerifier string) (*http.Request, error) {
	tokenURL := fmt.Sprintf("%s/auth/oauth/token", baseAuthURL)
	payload := url.Values{
		"grant_type":             []string{"authorization_code"},
		"code":                   []string{code},
		"client_id":              []string{clientID},
		"redirect_uri":           []string{redirectURI},
		"code_verifier":          []string{codeVerifier},
		"token_response_version": []string{"standard"}, // For backward compatibility with older Rill CLI, see utils.go in oauth pkg
	}
	req, err := http.NewRequest(
		http.MethodPost,
		tokenURL,
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", oauth.FormMediaType)
	req.Header.Set("Accept", oauth.JSONMediaType)
	return req, nil
}

// generateCodeVerifier creates a cryptographically secure random string
// which is between 43 and 128 characters long using the specified charset.
func generateCodeVerifier() (string, error) {
	// Generate a random number between 0 and 85 to extend the length of the code verifier
	r, err := rand.Int(rand.Reader, big.NewInt(86))
	if err != nil {
		return "", err
	}
	// Define the length of the code verifier
	// Here, we randomly choose a length between 43 and 128 characters
	n := 43 + int(r.Int64())

	// Create a byte slice of length n to store the characters of our code verifier
	b := make([]byte, n)
	// Temp slice to read random numbers into
	temp := make([]byte, n)
	if _, err := rand.Read(temp); err != nil {
		return "", err
	}

	// Assign a valid character from charset for each byte in b
	for i := 0; i < n; i++ {
		b[i] = charset[temp[i]%byte(len(charset))]
	}

	return string(b), nil
}

// createCodeChallenge takes a codeVerifier and returns its SHA256 hash
// encoded in Base64 URL encoding without padding, which is the code challenge.
func createCodeChallenge(codeVerifier string) string {
	// Create a new SHA256 hash instance
	hasher := sha256.New()

	// Write the codeVerifier to the hasher
	hasher.Write([]byte(codeVerifier))

	// Compute the SHA256 hash
	hash := hasher.Sum(nil)

	// Base64 URL encode the hash
	return base64.RawURLEncoding.EncodeToString(hash)
}
