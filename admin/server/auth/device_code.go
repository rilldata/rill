package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/oauth"
)

// DeviceCodeResponse encapsulates the response for obtaining a device code.
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationCompleteURI string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	PollingInterval         int    `json:"interval"`
}

// TokenRequest encapsulates the request for obtaining an access token.
type TokenRequest struct {
	GrantType  string `json:"grant_type"`
	DeviceCode string `json:"device_code"`
	ClientID   string `json:"client_id"`
}

// handleDeviceCodeRequest creates a 24 digit random device code and 8 digit user code and returns that
// to the client. The device code is used to poll for an access token, and the user code is displayed
// to the user and is used to authorize the device. The device code and user code are stored in the
// server's device code store.
func (a *Authenticator) handleDeviceCodeRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to read request body: %w", err))
		return
	}
	bodyStr := string(body)
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	clientID := values.Get("client_id")
	if clientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}
	scopes := strings.Split(values.Get("scope"), " ")
	if len(scopes) == 0 {
		http.Error(w, "scope is required", http.StatusBadRequest)
		return
	}
	if len(scopes) > 1 || scopes[0] != "full_account" {
		http.Error(w, "invalid scope", http.StatusBadRequest)
		return
	}
	authCode, err := a.admin.IssueDeviceAuthCode(r.Context(), clientID)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to issue auth code: %w", err))
		return
	}

	// add a "-" after the 4th character
	readableUserCode := authCode.UserCode[:4] + "-" + authCode.UserCode[4:]

	qry := map[string]string{"user_code": readableUserCode}
	if values.Get("redirect") != "" {
		qry["redirect"] = values.Get("redirect")
	} else {
		qry["redirect"] = a.admin.URLs.AuthCLISuccessUI()
	}

	resp := DeviceCodeResponse{
		DeviceCode:              authCode.DeviceCode,
		UserCode:                readableUserCode,
		VerificationURI:         a.admin.URLs.AuthVerifyDeviceUI(nil),
		VerificationCompleteURI: a.admin.URLs.AuthVerifyDeviceUI(qry),
		ExpiresIn:               int(admin.DeviceAuthCodeTTL.Seconds()),
		PollingInterval:         5,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to write response, %w", err))
		return
	}
}

// handleUserCodeConfirmation handles the user code confirmation page. The user code is displayed
// to the user and they are asked to confirm that they want to authorize the device. If the user
// confirms, the device code is marked as approved in the server's device code store.
func (a *Authenticator) handleUserCodeConfirmation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}
	userCode := r.URL.Query().Get("user_code")
	if userCode == "" {
		http.Error(w, "user_code is required", http.StatusBadRequest)
		return
	}
	confirmation := r.URL.Query().Get("code_confirmed")
	if confirmation == "" {
		http.Error(w, "no code confirmation", http.StatusBadRequest)
		return
	}

	claims := GetClaims(r.Context())
	if claims == nil {
		internalServerError(w, fmt.Errorf("did not find any claims, %w", errors.New("server error")))
		return
	}
	if claims.OwnerType() != OwnerTypeUser {
		http.Error(w, "only users can confirm device codes", http.StatusBadRequest)
		return
	}
	userID := claims.OwnerID()

	// Remove "-" from user code
	userCode = strings.ReplaceAll(userCode, "-", "")
	authCode, err := a.admin.DB.FindPendingDeviceAuthCodeByUserCode(r.Context(), userCode)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, fmt.Sprintf("no such user code: %s found", userCode), http.StatusBadRequest)
			return
		}
		internalServerError(w, fmt.Errorf("failed to get auth code for userCode: %s, %w", userCode, err))
		return
	}

	// Update user code with user id and approval
	authCode.UserID = &userID
	if confirmation != "true" {
		authCode.ApprovalState = database.DeviceAuthCodeStateRejected
	} else {
		authCode.ApprovalState = database.DeviceAuthCodeStateApproved
	}
	err = a.admin.DB.UpdateDeviceAuthCode(r.Context(), authCode.ID, userID, authCode.ApprovalState)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to update auth code for userCode: %s, %w", userCode, err))
	}
}

// getAccessTokenForDeviceCode verifies the device code and returns an access token if the request is approved
func (a *Authenticator) getAccessTokenForDeviceCode(w http.ResponseWriter, r *http.Request, values url.Values) {
	deviceCode := values.Get("device_code")
	if deviceCode == "" {
		http.Error(w, "device_code is required", http.StatusBadRequest)
		return
	}
	grantType := values.Get("grant_type")
	if grantType != deviceCodeGrantType {
		http.Error(w, "invalid grant_type", http.StatusBadRequest)
		return
	}
	responseVersion := values.Get("token_response_version")

	authCode, err := a.admin.DB.FindDeviceAuthCodeByDeviceCode(r.Context(), deviceCode)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, fmt.Sprintf("no such device code: %s found", deviceCode), http.StatusBadRequest)
			return
		}
		internalServerError(w, fmt.Errorf("failed to get auth code for deviceCode: %s, %w", deviceCode, err))
		return
	}
	clientID := values.Get("client_id")
	if clientID != authCode.ClientID {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return
	}

	if authCode.Expiry.Before(time.Now()) {
		http.Error(w, "expired_token", http.StatusUnauthorized)
		return
	}
	if authCode.ApprovalState == database.DeviceAuthCodeStateRejected {
		err = a.admin.DB.DeleteDeviceAuthCode(r.Context(), deviceCode)
		if err != nil {
			internalServerError(w, fmt.Errorf("failed to clean up rejected code: %s, %w", deviceCode, err))
			return
		}
		http.Error(w, "rejected", http.StatusUnauthorized)
		return
	}
	if authCode.ApprovalState == database.DeviceAuthCodeStatePending {
		http.Error(w, "authorization_pending", http.StatusUnauthorized)
		return
	}
	if authCode.ApprovalState != database.DeviceAuthCodeStateApproved || authCode.UserID == nil {
		internalServerError(w, fmt.Errorf("inconsistent state, %w", errors.New("server error")))
		return
	}
	// TODO handle too many requests

	authToken, err := a.admin.IssueUserAuthToken(r.Context(), *authCode.UserID, authCode.ClientID, "", nil, nil, false)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to issue access token, %w", err))
		return
	}

	err = a.admin.DB.DeleteDeviceAuthCode(r.Context(), deviceCode)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to clean up approved code: %s, %w", deviceCode, err))
		return
	}

	var respBytes []byte
	if responseVersion == "standard" {
		resp := oauth.TokenResponse{
			AccessToken: authToken.Token().String(),
			TokenType:   "Bearer",
			ExpiresIn:   0, // never expires
			UserID:      *authCode.UserID,
		}
		respBytes, err = json.Marshal(resp)
	} else {
		resp := oauth.LegacyTokenResponse{
			AccessToken: authToken.Token().String(),
			TokenType:   "Bearer",
			ExpiresIn:   0, // never expires
			UserID:      *authCode.UserID,
		}
		respBytes, err = json.Marshal(resp)
	}

	if err != nil {
		internalServerError(w, fmt.Errorf("failed to marshal response, %w", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to write response, %w", err))
		return
	}
}

func internalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
