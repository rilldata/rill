package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/rilldata/rill/admin/database"
)

const TokenExpirationSeconds = 60 * 10 // 10 minutes =

func (s *Service) IssueAuthCode(ctx context.Context, clientID string) (*database.AuthCode, error) {
	authCode, err := generateDeviceAndUserCode()
	if err != nil {
		return nil, err
	}
	authCode.ClientID = clientID
	err = s.DB.CreateAuthCode(ctx, authCode)
	if err != nil {
		return nil, err
	}
	return authCode, nil
}

// generateDeviceAndUserCode generates a random device code and user code.
func generateDeviceAndUserCode() (*database.AuthCode, error) {
	// Generate a random 24-byte device code, after base64 encoding it will be 32 characters
	deviceCodeBytes := make([]byte, 24)
	_, err := rand.Read(deviceCodeBytes)
	if err != nil {
		return nil, err
	}
	deviceCode := base64.StdEncoding.EncodeToString(deviceCodeBytes)

	// Generate an 8-character base 36 user code from the device code
	base36Chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	userCode := ""
	for i := 0; i < 8; i++ {
		b := deviceCodeBytes[i*3]
		userCode += base36Chars[b%36 : b%36+1]
	}

	return &database.AuthCode{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		Expiry:     time.Now().Add(TokenExpirationSeconds * time.Second),
	}, nil
}
