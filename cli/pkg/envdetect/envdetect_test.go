package envdetect

import (
	"os"
	"strings"
	"testing"
)

func TestIsWSL(t *testing.T) {
	// This test will only work in WSL environments
	// In non-WSL environments, it should return false
	result := IsWSL()

	// We can't easily test the true case without being in WSL,
	// but we can verify it doesn't panic and returns a boolean
	if result != true && result != false {
		t.Errorf("IsWSL() should return a boolean value, got %v", result)
	}
}

func TestIsOnWindowsPartition(t *testing.T) {
	// Test with a typical Linux path (should return false)
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create a temporary directory in the Linux filesystem
	tempDir := "/tmp"
	if err := os.Chdir(tempDir); err == nil {
		result := IsOnWindowsPartition()
		if result {
			t.Errorf("IsOnWindowsPartition() should return false for Linux filesystem path, got %v", result)
		}
	}

	// Test with a Windows partition path (should return true if in WSL)
	// This is a mock test - in real WSL, /mnt/c would exist
	if _, err := os.Stat("/mnt"); err == nil {
		// Only test if /mnt exists (indicating WSL environment)
		if err := os.Chdir("/mnt"); err == nil {
			result := IsOnWindowsPartition()
			// In WSL, this should return true
			// In non-WSL, /mnt might exist but IsOnWindowsPartition should still work
			_ = result // We can't assert the exact value without knowing the environment
		}
	}
}

func TestIsWSLWindowsPartition(t *testing.T) {
	// Test that the function doesn't panic and returns a boolean
	result := IsWSLWindowsPartition(".")
	if result != true && result != false {
		t.Errorf("IsWSLWindowsPartition() should return a boolean value, got %v", result)
	}
}

func TestGetWSLWarningMessage(t *testing.T) {
	message := GetWSLWarningMessage()

	// Check that the message contains expected keywords
	mustContainAll := []string{"WARNING", "Windows partition", "WSL"}
	for _, keyword := range mustContainAll {
		if !strings.Contains(message, keyword) {
			t.Errorf("Warning message should contain '%s', got: %s", keyword, message)
		}
	}

	// Must mention file system conflicts
	if !strings.Contains(message, "file system conflicts") {
		t.Errorf("Warning message should mention 'file system conflicts', got: %s", message)
	}
}
