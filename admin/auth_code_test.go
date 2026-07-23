package admin

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestGenerateDeviceAndUserCode(t *testing.T) {
	// Check the externally visible code format and lifetime. Collision handling is
	// covered deterministically at the database-backed issuance boundary.
	started := time.Now()
	authCode, err := generateDeviceAndUserCode()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(authCode.DeviceCode)
	if err != nil {
		t.Fatalf("device code is not valid base64: %v", err)
	}
	if len(decoded) != 24 {
		t.Errorf("decoded device code length = %d, want 24", len(decoded))
	}
	if len(authCode.UserCode) != 8 || strings.Trim(authCode.UserCode, "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		t.Errorf("user code %q is not eight uppercase base36 characters", authCode.UserCode)
	}
	if authCode.Expiry.Before(started.Add(DeviceAuthCodeTTL)) || authCode.Expiry.After(time.Now().Add(DeviceAuthCodeTTL)) {
		t.Errorf("expiry %s is outside the issuance window", authCode.Expiry)
	}
}
