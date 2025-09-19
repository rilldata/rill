package cmdutil

import (
	"os"
	"runtime"
	"strings"
)

// IsWSL detects if the application is running in WSL (Windows Subsystem for Linux)
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	
	// Check for WSL-specific indicators
	// Method 1: Check for WSL in /proc/version
	if version, err := os.ReadFile("/proc/version"); err == nil {
		versionStr := strings.ToLower(string(version))
		if strings.Contains(versionStr, "microsoft") || strings.Contains(versionStr, "wsl") {
			return true
		}
	}
	
	// Method 2: Check for WSL environment variable
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	
	// Method 3: Check for WSL in /proc/sys/kernel/osrelease
	if osrelease, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		osreleaseStr := strings.ToLower(string(osrelease))
		if strings.Contains(osreleaseStr, "microsoft") || strings.Contains(osreleaseStr, "wsl") {
			return true
		}
	}
	
	return false
}

// IsWindowsPathInWSL checks if the current directory is a Windows path when running in WSL
func IsWindowsPathInWSL(currentDir string) bool {
	if !IsWSL() {
		return false
	}
	
	// Check if the path starts with /mnt/ (WSL mount point for Windows drives)
	if strings.HasPrefix(currentDir, "/mnt/") {
		return true
	}
	
	// Check for Windows-style drive letters (C:, D:, etc.)
	if len(currentDir) >= 3 && currentDir[1] == ':' && currentDir[2] == '/' {
		return true
	}
	
	// Check for Windows-style paths that might be accessed through WSL
	// This catches cases where users might be in a Windows directory
	// that's been mounted or accessed through WSL
	windowsDrivePatterns := []string{
		"/mnt/c/", "/mnt/d/", "/mnt/e/", "/mnt/f/", "/mnt/g/", "/mnt/h/",
		"/mnt/i/", "/mnt/j/", "/mnt/k/", "/mnt/l/", "/mnt/m/", "/mnt/n/",
		"/mnt/o/", "/mnt/p/", "/mnt/q/", "/mnt/r/", "/mnt/s/", "/mnt/t/",
		"/mnt/u/", "/mnt/v/", "/mnt/w/", "/mnt/x/", "/mnt/y/", "/mnt/z/",
	}
	
	for _, pattern := range windowsDrivePatterns {
		if strings.HasPrefix(strings.ToLower(currentDir), pattern) {
			return true
		}
	}
	
	return false
}

// GetWSLGuidanceMessage returns a helpful message for WSL users who are in the wrong directory
func GetWSLGuidanceMessage(currentDir string) string {
	// Try to suggest the WSL equivalent path
	wslPath := currentDir
	
	// Convert Windows path to WSL path if possible
	if strings.HasPrefix(currentDir, "/mnt/") {
		// Extract drive letter and path from /mnt/c/Users/... format
		parts := strings.SplitN(currentDir, "/", 4)
		if len(parts) >= 4 {
			driveLetter := strings.ToLower(parts[2])
			windowsPath := "/" + parts[3]
			wslPath = "/mnt/" + driveLetter + windowsPath
		}
	}
	
	return `You appear to be running Rill from a Windows directory while in WSL.

This can cause issues with file permissions and path handling. For the best experience, please:

1. Navigate to your WSL home directory:
   cd ~

2. Or navigate to a WSL directory:
   cd /home/yourusername/your-project

3. Then run rill start again.

Current directory: ` + currentDir + `
Suggested WSL directory: ` + wslPath + `

Would you like to continue anyway? (This may cause issues)`
}

// GetWSLHomeDirectory returns the WSL home directory path
func GetWSLHomeDirectory() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	return "/home/" + os.Getenv("USER")
}
