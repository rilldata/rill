package oauth

const (
	FormMediaType = "application/x-www-form-urlencoded"
	JSONMediaType = "application/json"
)

// TokenResponse contains the information returned after fetching an access token from the OAuth server.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in,string"`
	TokenType   string `json:"token_type"`
	UserID      string `json:"user_id"`
}
