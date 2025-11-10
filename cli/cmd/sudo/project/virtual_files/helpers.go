package virtual_files

import (
	"context"
	"fmt"
	"strings"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
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

// pullVirtualFiles retrieves all virtual files for a given project, handling pagination
func pullVirtualFiles(ctx context.Context, client adminv1.AdminServiceClient, projectID, environment string, pageSize uint32) ([]*adminv1.VirtualFile, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	var allFiles []*adminv1.VirtualFile
	pageToken := ""

	for {
		res, err := client.PullVirtualRepo(ctx, &adminv1.PullVirtualRepoRequest{
			ProjectId:   projectID,
			Environment: environment,
			PageSize:    pageSize,
			PageToken:   pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to pull virtual repo: %w", err)
		}

		if res.Files == nil {
			break
		}

		allFiles = append(allFiles, res.Files...)

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	return allFiles, nil
}
