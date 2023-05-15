package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
)

const DeviceAuthCodeTTL = 10 * time.Minute

func (s *Service) IssueDeviceAuthCode(ctx context.Context, clientID string) (*database.DeviceAuthCode, error) {
	authCode, err := generateDeviceAndUserCode()
	if err != nil {
		return nil, err
	}
	authCode.ClientID = clientID
	code, err := s.DB.InsertDeviceAuthCode(ctx, authCode.DeviceCode, authCode.UserCode, authCode.ClientID, authCode.Expiry)
	if err != nil {
		return nil, err
	}
	return code, nil
}

// generateDeviceAndUserCode generates a random device code and user code.
func generateDeviceAndUserCode() (*database.DeviceAuthCode, error) {
	// Generate a random 24-byte device code, after base64 encoding it will be 32 characters
	deviceCodeBytes := make([]byte, 24)
	_, err := rand.Read(deviceCodeBytes)
	if err != nil {
		return nil, err
	}
	deviceCode := base64.StdEncoding.EncodeToString(deviceCodeBytes)

	userCode, err := generateUserCode()
	if err != nil {
		return nil, err
	}

	return &database.DeviceAuthCode{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		Expiry:     time.Now().Add(DeviceAuthCodeTTL),
	}, nil
}

func generateUserCode() (string, error) {
	// Generate an 8-character base 36 user code from the device code
	userCodeBytes := make([]byte, 8)
	_, err := rand.Read(userCodeBytes)
	if err != nil {
		return "", err
	}
	var i big.Int
	i.SetBytes(userCodeBytes)
	userCode := strings.ToUpper(i.Text(36))
	if len(userCode) < 8 {
		userCode = strings.Repeat("0", 8-len(userCode)) + userCode
	} else if len(userCode) > 8 {
		userCode = userCode[:8]
	}
	return userCode, nil
}
