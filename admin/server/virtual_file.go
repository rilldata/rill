package server

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin"
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

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	owner, err := personalFileOwnerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	var virtualFileNames []string
	virtualFiles, err := s.findPersonalFiles(ctx, proj.ID, owner)
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

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	owner, err := personalFileOwnerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	// Embed users are untrusted outside users, so we cap how many personal files they can create.
	if owner.embedUserID != "" {
		existing, err := s.findPersonalFiles(ctx, proj.ID, owner)
		if err != nil {
			return nil, fmt.Errorf("failed to list personal files: %w", err)
		}
		if len(existing) >= maxPersonalFilesPerEmbedUser {
			return nil, status.Errorf(codes.ResourceExhausted, "cannot create more than %d personal files", maxPersonalFilesPerEmbedUser)
		}
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

	virtualPath := virtualFilePathForPersonalFile(name)

	yaml, err := yamlForPersonalFile(req.DisplayName, owner, req.Kind, req.Yaml)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:    proj.ID,
		Environment:  "prod",
		Path:         virtualPath,
		OwnerID:      owner.userIDPtr(),
		EmbedOwnerID: owner.embedUserIDPtr(),
		Data:         yaml,
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

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	owner, err := personalFileOwnerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	if !owner.owns(vf) {
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

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	owner, err := personalFileOwnerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	if !owner.owns(vf) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update file")
	}

	// TODO: display name can be changed from yaml, so we dont need it here
	yaml, err := yamlForPersonalFile("", owner, req.Kind, req.Yaml)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:    proj.ID,
		Environment:  "prod",
		Path:         virtualPath,
		OwnerID:      owner.userIDPtr(),
		EmbedOwnerID: owner.embedUserIDPtr(),
		Data:         yaml,
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

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	owner, err := personalFileOwnerFromClaims(claims)
	if err != nil {
		return nil, err
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, err
	}

	virtualPath := virtualFilePathForPersonalFile(req.Name)

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", virtualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal file: %w", err)
	}

	if !owner.owns(vf) {
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

// findPersonalFiles returns the non-deleted personal files owned by the owner in the project's prod environment.
func (s *Server) findPersonalFiles(ctx context.Context, projectID string, owner *personalFileOwner) ([]*database.VirtualFile, error) {
	if owner.embedUserID != "" {
		return s.admin.DB.FindVirtualFilesByEmbedOwner(ctx, projectID, "prod", owner.embedUserID)
	}
	return s.admin.DB.FindVirtualFilesByOwner(ctx, projectID, "prod", owner.userID)
}

// maxPersonalFilesPerEmbedUser caps the number of live personal files an embed user can have in a project.
const maxPersonalFilesPerEmbedUser = 50

// personalFileOwner identifies the owner of a personal file.
// Exactly one of userID and embedUserID is set.
type personalFileOwner struct {
	userID      string
	embedUserID string
}

// personalFileOwnerFromClaims resolves the personal file owner for the caller.
// Only Rill users and embed users can own personal files.
func personalFileOwnerFromClaims(claims auth.Claims) (*personalFileOwner, error) {
	switch claims.OwnerType() {
	case auth.OwnerTypeUser:
		return &personalFileOwner{userID: claims.OwnerID()}, nil
	case auth.OwnerTypeEmbed:
		// Embeds with fixed resources issue runtime JWTs that only grant access to those resources,
		// so a personal canvas created from them could never be opened.
		tkn, ok := claims.AuthTokenModel().(*admin.EmbedToken)
		if !ok {
			return nil, status.Error(codes.Internal, "unexpected embed token model")
		}
		if len(tkn.Resources) > 0 {
			return nil, status.Error(codes.FailedPrecondition, "personal files are not available for embeds limited to specific resources")
		}
		return &personalFileOwner{embedUserID: claims.OwnerID()}, nil
	default:
		return nil, status.Error(codes.PermissionDenied, "only users can manage personal files")
	}
}

// owns returns true if the virtual file is owned by the owner.
func (o *personalFileOwner) owns(vf *database.VirtualFile) bool {
	if o.embedUserID != "" {
		return vf.EmbedOwnerID != nil && *vf.EmbedOwnerID == o.embedUserID
	}
	return vf.OwnerID != nil && *vf.OwnerID == o.userID
}

func (o *personalFileOwner) userIDPtr() *string {
	if o.userID == "" {
		return nil
	}
	return &o.userID
}

func (o *personalFileOwner) embedUserIDPtr() *string {
	if o.embedUserID == "" {
		return nil
	}
	return &o.embedUserID
}

// personalFileBlockedKeys are top-level properties that personal files may not set.
// They either change the resource's identity (name, kind) or are irrelevant for canvases.
// The environment override keys are blocked because the runtime parser applies them on top of the stored YAML,
// which would bypass the type and annotation enforcement in yamlForPersonalFile.
var personalFileBlockedKeys = map[string]bool{
	"env":                   true,
	"environment_overrides": true,
	"dev":                   true,
	"prod":                  true,
	"name":                  true,
	"namespace":             true,
	"kind":                  true,
	"refs":                  true,
	"connector":             true,
	"sql":                   true,
}

// yamlForPersonalFile builds the YAML stored for a personal file.
// If data is empty, it creates a blank file.
// It strips any client-supplied admin_* annotations and sets the server-controlled ones for the owner.
func yamlForPersonalFile(displayName string, owner *personalFileOwner, kind, data string) ([]byte, error) {
	if kind != runtime.ResourceKindCanvas {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal file kind: %s", kind)
	}

	var doc map[string]any
	if err := yaml.Unmarshal([]byte(data), &doc); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid YAML: %s", err.Error())
	}
	if doc == nil {
		doc = map[string]any{"rows": []any{}}
	}

	for k := range doc {
		if personalFileBlockedKeys[k] {
			return nil, status.Errorf(codes.InvalidArgument, "property %q is not allowed in personal files", k)
		}
	}

	if displayName != "" {
		doc["display_name"] = displayName
	}

	// Hardcode the kind to prevent client from changing it
	doc["type"] = "canvas"

	annotations, _ := doc["annotations"].(map[string]any)
	sanitized := map[string]any{}
	for k, v := range annotations {
		if !strings.HasPrefix(k, "admin_") {
			sanitized[k] = v
		}
	}
	if owner.embedUserID != "" {
		// Sharing is not supported for embed-owned files.
		sanitized["admin_embed_owner_id"] = owner.embedUserID
	} else {
		if shared, ok := annotations["admin_shared"]; ok {
			sanitized["admin_shared"] = shared
		}
		sanitized["admin_owner_user_id"] = owner.userID
	}
	sanitized["admin_managed"] = "true"
	sanitized["admin_nonce"] = time.Now().Format(time.RFC3339Nano)
	doc["annotations"] = sanitized

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal YAML: %s", err.Error())
	}
	return out, nil
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
