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
	"github.com/rilldata/rill/admin/pkg/urlutil"
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

	// Only support creating canvas resource for now. Should be a very narrow allow list for security.
	if req.Kind != runtime.ResourceKindCanvas {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal file kind %q", req.Kind)
	}

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
		_, err := s.admin.LookupCanvas(ctx, depl, name)
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

	userID := auth.GetClaims(ctx).OwnerID()
	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	if vf.OwnerID != nil && *vf.OwnerID != userID {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read file")
	}

	return &adminv1.GetPersonalFileResponse{
		Path: path.Join("__virtual__", virtualPath),
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

	// Only support creating canvas resource for now. Should be a very narrow allow list for security.
	if req.Kind != runtime.ResourceKindCanvas {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal file kind %q", req.Kind)
	}

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
	if err != nil {
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
	if err != nil {
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

func (s *Server) SavePersonalFileForUser(ctx context.Context, req *adminv1.SavePersonalFileForUserRequest) (*adminv1.SavePersonalFileForUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.String("args.user_id", req.UserId),
		attribute.String("args.kind", req.Kind),
		attribute.String("args.name", req.Name),
	)

	// Only support canvas resources for now. Should be a very narrow allow list for security.
	if req.Kind != runtime.ResourceKindCanvas {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal file kind %q", req.Kind)
	}
	if req.Yaml == "" {
		return nil, status.Error(codes.InvalidArgument, "yaml is required")
	}
	if req.Name == "" && req.DisplayName == "" {
		return nil, status.Error(codes.InvalidArgument, "display_name is required when creating a personal file")
	}
	// The name comes from an AI tool call, so only accept the alphabet generateVirtualFileName produces.
	if req.Name != "" && !personalFileNameRegexp.MatchString(req.Name) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid personal file name %q", req.Name)
	}

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	// Only the project's primary deployment may act on behalf of a user.
	// It is trusted to pass the user of its session, as for GetReportMeta and GetAlertMeta.
	// Other deployments are rejected because personal files live in the prod environment, which only the primary deployment reads.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeDeployment {
		return nil, status.Error(codes.PermissionDenied, "only a deployment can save personal files for a user")
	}
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}
	if proj.PrimaryDeploymentID == nil || *proj.PrimaryDeploymentID != claims.OwnerID() {
		return nil, status.Error(codes.PermissionDenied, "only the project's primary deployment can save personal files for a user")
	}

	// The owner must be a user (not a service, magic token or external embed user) with access to the project.
	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, status.Error(codes.PermissionDenied, "personal files can only be owned by a user")
	}
	user, err := s.admin.DB.FindUser(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.PermissionDenied, "personal files can only be owned by a user")
		}
		return nil, err
	}
	orgPerms, err := s.admin.OrganizationPermissionsForUser(ctx, proj.OrganizationID, user.ID)
	if err != nil {
		return nil, err
	}
	projPerms, err := s.admin.ProjectPermissionsForUser(ctx, proj.ID, user.ID, orgPerms)
	if err != nil {
		return nil, err
	}
	if !projPerms.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "user does not have permission to read project")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	name := req.Name
	if name == "" {
		name, err = s.createPersonalFileForUser(ctx, proj, depl, user.ID, req)
	} else {
		err = s.editPersonalFileForUser(ctx, proj, user.ID, req)
	}
	if err != nil {
		return nil, err
	}

	// Don't wait for the resource: the caller polls the runtime itself, which also surfaces parse errors in the YAML.
	err = s.admin.TriggerParser(ctx, depl)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger parser: %w", err)
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	return &adminv1.SavePersonalFileForUserResponse{
		Name: name,
		Url:  urlutil.MustJoinURL(s.admin.URLs.WithCustomDomain(org.CustomDomain).Project(org.Name, proj.Name), "canvas", name),
	}, nil
}

// createPersonalFileForUser creates a personal file owned by userID with a new name, and returns the name.
func (s *Server) createPersonalFileForUser(ctx context.Context, proj *database.Project, depl *database.Deployment, userID string, req *adminv1.SavePersonalFileForUserRequest) (string, error) {
	name, err := s.generateVirtualFileName(ctx, req.DisplayName, func(ctx context.Context, name string) error {
		_, err := s.admin.LookupCanvas(ctx, depl, name)
		return err
	})
	if err != nil {
		return "", err
	}

	data, err := yamlForPersonalFileOnBehalf(req.Yaml, false)
	if err != nil {
		return "", err
	}
	yaml, err := yamlForPersonalFile(req.DisplayName, userID, req.Kind, string(data))
	if err != nil {
		return "", err
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        virtualFilePathForPersonalFile(name),
		OwnerID:     &userID,
		Data:        yaml,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save personal file: %w", err)
	}
	return name, nil
}

// editPersonalFileForUser overwrites a personal file owned by userID.
// It keeps the file's sharing state, since sharing is the user's decision and not the caller's,
// and it fails instead of overwriting if the user shared, unshared or deleted the file concurrently.
// The display name comes from the YAML, as in EditPersonalFile.
func (s *Server) editPersonalFileForUser(ctx context.Context, proj *database.Project, userID string, req *adminv1.SavePersonalFileForUserRequest) error {
	path := virtualFilePathForPersonalFile(req.Name)
	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", path)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return status.Errorf(codes.NotFound, "personal file %q not found", req.Name)
		}
		return fmt.Errorf("failed to get personal file: %w", err)
	}
	if vf.Deleted {
		return status.Errorf(codes.NotFound, "personal file %q not found", req.Name)
	}
	if vf.OwnerID == nil || *vf.OwnerID != userID {
		return status.Error(codes.PermissionDenied, "does not have permission to update file")
	}

	data, err := yamlForPersonalFileOnBehalf(req.Yaml, personalFileShared(vf.Data))
	if err != nil {
		return err
	}
	yaml, err := yamlForPersonalFile("", userID, req.Kind, string(data))
	if err != nil {
		return err
	}

	err = s.admin.DB.UpdateVirtualFileDataIfUnchanged(ctx, &database.UpdateVirtualFileDataOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        path,
		Data:        yaml,
		UpdatedOn:   vf.UpdatedOn,
	})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return status.Errorf(codes.Aborted, "personal file %q changed while it was being saved, try again", req.Name)
		}
		return fmt.Errorf("failed to save personal file: %w", err)
	}
	return nil
}

// yamlForPersonalFileOnBehalf drops the annotations of a personal file written on behalf of a user,
// so the caller can neither set the owner nor share the file. It sets admin_shared to the file's current sharing state.
func yamlForPersonalFileOnBehalf(data string, shared bool) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(data), &doc); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid YAML: %s", err.Error())
	}
	if doc == nil {
		return nil, status.Error(codes.InvalidArgument, "yaml is empty")
	}

	delete(doc, "annotations")
	if shared {
		doc["annotations"] = map[string]any{"admin_shared": "true"}
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal YAML: %s", err.Error())
	}
	return out, nil
}

// personalFileShared reports whether a personal file's YAML marks it as shared with the project.
func personalFileShared(data []byte) bool {
	var doc struct {
		Annotations map[string]any `yaml:"annotations"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false
	}
	return doc.Annotations["admin_shared"] == "true"
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
	if kind == runtime.ResourceKindCanvas {
		// Hardcode the kind to prevent client from changing it
		doc["type"] = "canvas"
		// Delete the legacy key to prevent client using that as an override
		_, hasKind := doc["kind"]
		if hasKind {
			delete(doc, "kind")
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unsupported personal file kind: %s", kind))
	}

	annotations, _ := doc["annotations"].(map[string]any)
	if annotations == nil {
		annotations = map[string]any{}
	}
	annotations["admin_owner_user_id"] = ownerID
	annotations["admin_managed"] = "true"
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
			"admin_managed":       "true",
			"admin_nonce":         time.Now().Format(time.RFC3339Nano),
		},
	}
	if kind == runtime.ResourceKindCanvas {
		doc["type"] = "canvas"
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

// personalFileNameRegexp matches the names generateVirtualFileName produces.
var personalFileNameRegexp = regexp.MustCompile(`^[a-z0-9-]+$`)

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
