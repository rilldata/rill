package server

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

func (s *Server) ListPersonalFiles(ctx context.Context, req *adminv1.ListPersonalFilesRequest) (*adminv1.ListPersonalFilesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	var virtualFileNames []string
	virtualFiles, err := s.admin.DB.FindVirtualFilesByOwner(ctx, proj.ID, "prod", auth.GetClaims(ctx).OwnerID())
	if err != nil {
		return nil, fmt.Errorf("failed to list personal files: %w", err)
	}
	for _, file := range virtualFiles {
		// Assumes the name is the stem of the path, we block updating name from yaml
		virtualFileNames = append(virtualFileNames, fileutil.Stem(file.Path))
	}

	return &adminv1.ListPersonalFilesResponse{
		Files: virtualFileNames,
	}, nil
}

func (s *Server) CreatePersonalFile(ctx context.Context, req *adminv1.CreatePersonalFileRequest) (*adminv1.CreatePersonalFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.kind", req.Kind),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	name, err := s.generateVirtualFileName(ctx, req.DisplayName, func(ctx context.Context, name string) error {
		var err error
		if req.Kind == runtime.ResourceKindCanvas {
			_, err = s.admin.LookupCanvas(ctx, depl, name)
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	userID := auth.GetClaims(ctx).OwnerID()
	virtualPath := virtualFilePathForPersonalFile(name)

	yaml, err := yamlForPersonalFile(req.DisplayName, auth.GetClaims(ctx).OwnerID(), req.Kind, req.Yaml)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        virtualPath,
		OwnerID:     &userID,
		Data:        yaml,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create personal file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, name, req.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger parser: %w", err)
	}

	return &adminv1.CreatePersonalFileResponse{
		Name: name,
	}, nil
}

func (s *Server) GetPersonalFile(ctx context.Context, req *adminv1.GetPersonalFileRequest) (*adminv1.GetPersonalFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	return &adminv1.GetPersonalFileResponse{
		Path: virtualPath,
		Yaml: string(vf.Data),
	}, nil
}

func (s *Server) EditPersonalFile(ctx context.Context, req *adminv1.EditPersonalFileRequest) (*adminv1.EditPersonalFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
		attribute.String("args.kind", req.Kind),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	userID := auth.GetClaims(ctx).OwnerID()
	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	if vf.OwnerID != nil && *vf.OwnerID != userID {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update file")
	}

	// TODO: display name can be changed from yaml, so we dont need it here
	yaml, err := yamlForPersonalFile("", auth.GetClaims(ctx).OwnerID(), req.Kind, req.Yaml)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        virtualPath,
		OwnerID:     &userID,
		Data:        yaml,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create personal file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, req.Name, req.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger parser: %w", err)
	}

	return &adminv1.EditPersonalFileResponse{}, nil
}

func (s *Server) DeletePersonalFile(ctx context.Context, req *adminv1.DeletePersonalFileRequest) (*adminv1.DeletePersonalFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.name", req.Name),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	ownerID := auth.GetClaims(ctx).OwnerID()
	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	if vf.OwnerID != nil && *vf.OwnerID != ownerID {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete file")
	}

	err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, "prod", virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to delete virtual file: %w", err)
	}

	// TODO: do we need to wait for reconcile to finish?
	err = s.admin.TriggerParser(ctx, depl)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile personal virtual file: %w", err)
	}

	return &adminv1.DeletePersonalFileResponse{}, nil
}

func yamlForPersonalFile(displayName, ownerID, kind, data string) ([]byte, error) {
	if data == "" {
		return blankYamlForPersonalFile(displayName, ownerID, kind)
	}

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(data), &doc); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid YAML: %s", err.Error())
	}
	if doc == nil {
		doc = map[string]any{}
	}

	if displayName != "" {
		doc["display_name"] = displayName
	}

	annotations, _ := doc["annotations"].(map[string]any)
	if annotations == nil {
		annotations = map[string]any{}
	}
	annotations["admin_owner_user_id"] = ownerID
	annotations["admin_managed"] = true
	annotations["admin_nonce"] = time.Now().Format(time.RFC3339Nano)
	doc["annotations"] = annotations

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal YAML: %s", err.Error())
	}
	return out, nil
}

func blankYamlForPersonalFile(displayName, ownerID, kind string) ([]byte, error) {
	doc := map[string]any{
		"type":         kind,
		"display_name": displayName,
		"annotations": map[string]any{
			"admin_owner_user_id": ownerID,
			"admin_managed":       true,
			"admin_nonce":         time.Now().Format(time.RFC3339Nano),
		},
	}
	if kind == runtime.ResourceKindCanvas {
		doc["rows"] = []any{}
	}

	return yaml.Marshal(doc)
}

// generateVirtualFileName generates a random virtual file name with the display name as a seed.
// Example: "My report!" -> "my-report-5b3f7e1a".
// It verifies that the name is not taken (the random component makes any collision unlikely, but we check to be sure).
func (s *Server) generateVirtualFileName(ctx context.Context, displayName string, lookup func(ctx context.Context, name string) error) (string, error) {
	for i := 0; i < 5; i++ {
		name := randomVirtualFileName(displayName)

		err := lookup(ctx, name)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
				// Success! Name isn't taken
				return name, nil
			}
			return "", fmt.Errorf("failed to check virtual file name: %w", err)
		}
	}

	// Fail-safe in case all names we tried were taken
	return uuid.New().String(), nil
}

var virtualFileNameToDashCharsRegexp = regexp.MustCompile(`[ _]+`)

var virtualFileNameExcludeCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

func randomVirtualFileName(displayName string) string {
	name := virtualFileNameToDashCharsRegexp.ReplaceAllString(displayName, "-")
	name = virtualFileNameExcludeCharsRegexp.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	name = strings.Trim(name, "-")
	if name == "" {
		name = uuid.New().String()
	} else {
		name = name + "-" + uuid.New().String()[0:8]
	}
	return name
}

func virtualFilePathForPersonalFile(name string) string {
	return path.Join("personal", name+".yaml")
}
