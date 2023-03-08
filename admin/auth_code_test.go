package admin

import (
	"testing"
)

func TestGenerateDeviceAndUserCode(t *testing.T) {
	authCode, err := generateDeviceAndUserCode()
	if err != nil {
		t.Error(err)
	}
	if len(authCode.DeviceCode) != 32 {
		t.Errorf("device code length is incorrect; got %d, want 32", len(authCode.DeviceCode))
	}
	if len(authCode.UserCode) != 8 {
		t.Errorf("user code length is incorrect; got %d, want 8", len(authCode.UserCode))
	}
}

func TestGenerateDeviceAndUserCodeCollision(t *testing.T) {
	deviceCodeSet := make(map[string]bool)
	userCodeSet := make(map[string]bool)

	for i := 0; i < 10000; i++ {
		authCode, err := generateDeviceAndUserCode()
		if err != nil {
			t.Error(err)
		}
		if len(authCode.DeviceCode) != 32 {
			t.Errorf("device code length is incorrect; got %d, want 32", len(authCode.DeviceCode))
		}
		if len(authCode.UserCode) != 8 {
			t.Errorf("user code length is incorrect; got %d, want 8", len(authCode.UserCode))
		}
		if deviceCodeSet[authCode.DeviceCode] {
			t.Errorf("device code collision: %s", authCode.DeviceCode)
		}
		if userCodeSet[authCode.UserCode] {
			t.Errorf("user code collision: %s", authCode.UserCode)
		}
		deviceCodeSet[authCode.DeviceCode] = true
		userCodeSet[authCode.UserCode] = true
	}
}
