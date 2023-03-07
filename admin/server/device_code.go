package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"go.uber.org/zap"
)

const (
	tokenExpirationSeconds = 60 * 10 // 10 minutes
	deviceCodeGrantType    = "urn:ietf:params:oauth:grant-type:device_code"
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
func (s *Server) handleDeviceCodeRequest(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("failed to read request body", zap.Error(err))
		internalServerError(w, err)
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
	authCode, err := s.admin.IssueAuthCode(r.Context(), clientID)
	if err != nil {
		s.logger.Error("failed to issue auth code", zap.Error(err))
		internalServerError(w, err)
		return
	}

	verificationURI := s.conf.DeviceVerificationHost + "/oauth/device"
	resp := DeviceCodeResponse{
		DeviceCode:              authCode.DeviceCode,
		UserCode:                authCode.UserCode,
		VerificationURI:         verificationURI,
		VerificationCompleteURI: verificationURI + "?user_code=" + authCode.UserCode,
		ExpiresIn:               tokenExpirationSeconds,
		PollingInterval:         5,
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	err = e.Encode(resp)
	if err != nil {
		s.logger.Error("failed to encode response", zap.Error(err))
		internalServerError(w, err)
	}
}

// handleUserCodeConfirmation handles the user code confirmation page. The user code is displayed
// to the user and they are asked to confirm that they want to authorize the device. If the user
// confirms, the device code is marked as approved in the server's device code store.
func (s *Server) handleUserCodeConfirmation(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
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

	claims := auth.GetClaims(r.Context())
	if claims == nil {
		s.logger.Error("did not find any claims")
		internalServerError(w, errors.New("server error"))
		return
	}
	userID := claims.OwnerID()
	if userID == "" {
		s.logger.Error("did not find user id in claims")
		internalServerError(w, errors.New("server error"))
		return
	}

	authCode, err := s.admin.DB.FindAuthCodeByUserCode(r.Context(), userCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no such user code: %s found", userCode), http.StatusBadRequest)
			return
		}
		s.logger.Error("failed to get auth code for userCode: "+userCode, zap.Error(err))
		internalServerError(w, err)
		return
	}
	if authCode.ApprovalState != database.Pending {
		http.Error(w, "device code already used", http.StatusBadRequest)
		return
	}
	if authCode.Expiry.Before(time.Now()) {
		http.Error(w, "device code expired", http.StatusBadRequest)
		return
	}

	// Update user code with user id and approval
	authCode.UserID = userID
	if confirmation != "true" {
		authCode.ApprovalState = database.Rejected
	} else {
		authCode.ApprovalState = database.Approved
	}
	err = s.admin.DB.UpdateAuthCode(r.Context(), userCode, userID, authCode.ApprovalState)
	if err != nil {
		s.logger.Error("failed to update auth code for userCode: "+userCode, zap.Error(err))
		internalServerError(w, err)
	}
}

// getAccessToken verifies the device code and returns an access token if the request is approved
func (s *Server) getAccessToken(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("failed to read request body", zap.Error(err))
		internalServerError(w, err)
		return
	}
	bodyStr := string(body)
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		s.logger.Error("failed to parse query: "+bodyStr, zap.Error(err))
		internalServerError(w, err)
		return
	}
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

	authCode, err := s.admin.DB.FindAuthCodeByDeviceCode(r.Context(), deviceCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, fmt.Sprintf("no such device code: %s found", deviceCode), http.StatusBadRequest)
			return
		}
		s.logger.Error("failed to get auth code for deviceCode: "+deviceCode, zap.Error(err))
		internalServerError(w, err)
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
	if authCode.ApprovalState == database.Rejected {
		http.Error(w, "rejected", http.StatusUnauthorized)
		return
	}
	if authCode.ApprovalState == database.Pending {
		http.Error(w, "authorization_pending", http.StatusUnauthorized)
		return
	}
	if authCode.ApprovalState != database.Approved || authCode.UserID == "" {
		s.logger.Error("inconsistent state")
		internalServerError(w, err)
		return
	}
	// TODO handle too many requests

	authToken, err := s.admin.IssueUserAuthToken(r.Context(), authCode.UserID, authCode.ClientID, "")
	if err != nil {
		s.logger.Error("failed to issue access token", zap.Error(err))
		internalServerError(w, err)
		return
	}

	// TODO insert access token and update auth code to used in a single transaction
	err = s.admin.DB.DeleteAuthCode(r.Context(), deviceCode)
	if err != nil {
		s.logger.Error("failed to update auth code to used for deviceCode: "+deviceCode, zap.Error(err))
		internalServerError(w, err)
		return
	}

	resp := deviceauth.OAuthTokenResponse{
		AccessToken: authToken.Token().String(),
		TokenType:   "Bearer",
		ExpiresIn:   time.UnixMilli(0).Unix(), // never expires
		UserID:      authCode.UserID,
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	err = e.Encode(resp)
	if err != nil {
		s.logger.Error("failed to encode response", zap.Error(err))
		internalServerError(w, err)
	}
}

func internalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
