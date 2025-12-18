package envdetect

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IsWSL returns true if the current environment is Windows Subsystem for Linux
func IsWSL() bool {
	// Check /proc/version for Microsoft signature
	if versionFile, err := os.Open("/proc/version"); err == nil {
		defer versionFile.Close()
		scanner := bufio.NewScanner(versionFile)
		if scanner.Scan() {
			version := scanner.Text()
			return strings.Contains(strings.ToLower(version), "microsoft")
		}
	}
	return false
}

// IsOnWindowsPartition returns true if the current working directory is on a Windows partition
// This is detected by checking if the path starts with /mnt/ (typical WSL mount point)
func IsOnWindowsPartition() bool {
	wd, err := os.Getwd()
	if err != nil {
		return false
	}
	return IsOnWindowsPartitionPath(wd)
}

// IsOnWindowsPartitionPath returns true if the provided path is on a Windows partition
// in WSL (typically paths mounted under /mnt/*)
func IsOnWindowsPartitionPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return strings.HasPrefix(absPath, "/mnt/")
}

// IsWSLWindowsPartition checks if the user is running on a Windows partition inside WSL
// Returns true if both conditions are met
func IsWSLWindowsPartition(path string) bool {
	return IsWSL() && IsOnWindowsPartitionPath(path)
}

// GetWSLWarningMessage returns the warning message for WSL Windows partition usage
func GetWSLWarningMessage() string {
	return "WARNING: You are running Rill on a Windows partition inside WSL. This is not recommended and will cause file system conflicts. Please run Rill from a Linux filesystem (e.g., ~/projects/my-rill-project) instead of a Windows drive (e.g., /mnt/c/...)."
}
