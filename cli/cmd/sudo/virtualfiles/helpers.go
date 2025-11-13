package virtualfiles

import (
	"strings"
)

const (
	FileTypeReport  string = "report"
	FileTypeAlert   string = "alert"
	FileTypeService string = "service"
	FileTypeUnknown string = "unknown"
)

// GetFileTypeAndName extracts the type and name from a virtual file path
func GetFileTypeAndName(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return FileTypeUnknown, ""
	}

	folder := parts[0]
	name := parts[1]

	// Strip .yaml extension if present
	name = strings.TrimSuffix(name, ".yaml")

	switch folder {
	case "reports":
		return FileTypeReport, name
	case "alerts":
		return FileTypeAlert, name
	case "services":
		return FileTypeService, name
	default:
		return FileTypeUnknown, ""
	}
}

// GetFileType extracts just the type from a path
func GetFileType(path string) string {
	fileType, _ := GetFileTypeAndName(path)
	return fileType
}
