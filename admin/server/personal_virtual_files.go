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
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// personalVirtualFileSpec describes how a personal virtual file type is laid out on disk and which
// permission is required to create/edit/delete one. Adding a new type (e.g. PERSONAL_REPORT) is a
// matter of adding one entry to personalVirtualFileSpecs below.
type personalVirtualFileSpec struct {
	// yamlType is the value of the `type:` field in the YAML body.
	yamlType string
	// runtimeKind is the runtime resource kind the parser produces for files of this type.
	runtimeKind string
	// pathPrefixSegment is the second segment in the path `personal/<segment>/<user_id>/<name>.yaml`.
	pathPrefixSegment string
	// hasPermission returns whether the caller has the permission required to manage personal files of this type.
	hasPermission func(*adminv1.ProjectPermissions) bool
	// buildBlankYAML returns a sensible default YAML body for a brand-new personal file.
	buildBlankYAML func(displayName, ownerUserID string) ([]byte, error)
}

var personalVirtualFileSpecs = map[adminv1.PersonalVirtualFileType]*personalVirtualFileSpec{
	adminv1.PersonalVirtualFileType_PERSONAL_VIRTUAL_FILE_TYPE_CANVAS: {
		yamlType:          "canvas",
		runtimeKind:       runtime.ResourceKindCanvas,
		pathPrefixSegment: "canvases",
		hasPermission: func(p *adminv1.ProjectPermissions) bool {
			return p.GetCreatePersonalCanvases()
		},
		buildBlankYAML: blankPersonalCanvasYAML,
	},
}

// CreatePersonalVirtualFile creates a personal (owner-only) virtual file for the calling user.
func (s *Server) CreatePersonalVirtualFile(ctx context.Context, req *adminv1.CreatePersonalVirtualFileRequest) (*adminv1.CreatePersonalVirtualFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.type", req.Type.String()),
		attribute.String("args.display_name", req.DisplayName),
	)

	spec, ok := personalVirtualFileSpecs[req.Type]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal virtual file type %q", req.Type.String())
	}

	proj, depl, claims, err := s.lookupProjectForPersonal(ctx, req.Org, req.Project, spec)
	if err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return nil, status.Error(codes.InvalidArgument, "display_name is required")
	}

	name := randomPersonalName(displayName)
	ownerID := claims.OwnerID()

	var data []byte
	if req.Yaml != "" {
		data, err = validatePersonalYAML(spec, []byte(req.Yaml), displayName, ownerID)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = spec.buildBlankYAML(displayName, ownerID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to build blank YAML: %s", err.Error())
		}
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        personalVirtualFilePath(spec, ownerID, name),
		Data:        data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, name, spec.runtimeKind)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile personal virtual file: %w", err)
	}

	return &adminv1.CreatePersonalVirtualFileResponse{Name: name}, nil
}

// EditPersonalVirtualFile replaces the YAML body of a personal virtual file the caller owns.
func (s *Server) EditPersonalVirtualFile(ctx context.Context, req *adminv1.EditPersonalVirtualFileRequest) (*adminv1.EditPersonalVirtualFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.type", req.Type.String()),
		attribute.String("args.name", req.Name),
	)

	spec, ok := personalVirtualFileSpecs[req.Type]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal virtual file type %q", req.Type.String())
	}

	proj, depl, claims, err := s.lookupProjectForPersonal(ctx, req.Org, req.Project, spec)
	if err != nil {
		return nil, err
	}

	existing, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", personalVirtualFilePath(spec, claims.OwnerID(), req.Name))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "personal virtual file not found")
		}
		return nil, err
	}
	if existing.Deleted {
		return nil, status.Error(codes.NotFound, "personal virtual file not found")
	}

	ownerID, _, err := parsePersonalAnnotations(existing.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse existing file annotations: %s", err.Error())
	}
	if ownerID != claims.OwnerID() {
		return nil, status.Error(codes.PermissionDenied, "only the owner can edit this personal virtual file")
	}

	data, err := validatePersonalYAML(spec, []byte(req.Yaml), "", ownerID)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        personalVirtualFilePath(spec, ownerID, req.Name),
		Data:        data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, req.Name, spec.runtimeKind)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile personal virtual file: %w", err)
	}

	return &adminv1.EditPersonalVirtualFileResponse{}, nil
}

// DeletePersonalVirtualFile soft-deletes a personal virtual file the caller owns.
func (s *Server) DeletePersonalVirtualFile(ctx context.Context, req *adminv1.DeletePersonalVirtualFileRequest) (*adminv1.DeletePersonalVirtualFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.type", req.Type.String()),
		attribute.String("args.name", req.Name),
	)

	spec, ok := personalVirtualFileSpecs[req.Type]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal virtual file type %q", req.Type.String())
	}

	proj, depl, claims, err := s.lookupProjectForPersonal(ctx, req.Org, req.Project, spec)
	if err != nil {
		return nil, err
	}

	pathStr := personalVirtualFilePath(spec, claims.OwnerID(), req.Name)
	existing, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", pathStr)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "personal virtual file not found")
		}
		return nil, err
	}
	if existing.Deleted {
		return nil, status.Error(codes.NotFound, "personal virtual file not found")
	}

	ownerID, _, err := parsePersonalAnnotations(existing.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse existing file annotations: %s", err.Error())
	}
	if ownerID != claims.OwnerID() {
		return nil, status.Error(codes.PermissionDenied, "only the owner can delete this personal virtual file")
	}

	err = s.admin.DB.UpdateVirtualFileDeleted(ctx, proj.ID, "prod", pathStr)
	if err != nil {
		return nil, fmt.Errorf("failed to delete virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, req.Name, spec.runtimeKind)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile personal virtual file: %w", err)
	}

	return &adminv1.DeletePersonalVirtualFileResponse{}, nil
}

// CopyPersonalVirtualFile clones a shared or own personal resource into a new personal virtual file.
func (s *Server) CopyPersonalVirtualFile(ctx context.Context, req *adminv1.CopyPersonalVirtualFileRequest) (*adminv1.CopyPersonalVirtualFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.type", req.Type.String()),
		attribute.String("args.source_kind", req.SourceKind.String()),
		attribute.String("args.source_name", req.SourceName),
	)

	spec, ok := personalVirtualFileSpecs[req.Type]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal virtual file type %q", req.Type.String())
	}

	proj, depl, claims, err := s.lookupProjectForPersonal(ctx, req.Org, req.Project, spec)
	if err != nil {
		return nil, err
	}

	ownerID := claims.OwnerID()
	displayName := strings.TrimSpace(req.DisplayName)

	var sourceData []byte
	var sourceDisplayName string

	switch req.SourceKind {
	case adminv1.PersonalVirtualFileSourceKind_PERSONAL_VIRTUAL_FILE_SOURCE_KIND_SHARED:
		// Read the shared resource from the runtime catalog and regenerate YAML.
		sourceData, sourceDisplayName, err = s.fetchSharedSourceYAML(ctx, depl, spec, req.SourceName, ownerID)
		if err != nil {
			return nil, err
		}
	case adminv1.PersonalVirtualFileSourceKind_PERSONAL_VIRTUAL_FILE_SOURCE_KIND_PERSONAL:
		// Read the source virtual file and verify ownership.
		src, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", personalVirtualFilePath(spec, ownerID, req.SourceName))
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, status.Error(codes.NotFound, "source personal virtual file not found")
			}
			return nil, err
		}
		if src.Deleted {
			return nil, status.Error(codes.NotFound, "source personal virtual file not found")
		}
		srcOwnerID, srcDisplayName, err := parsePersonalAnnotations(src.Data)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse source file annotations: %s", err.Error())
		}
		if srcOwnerID != ownerID {
			return nil, status.Error(codes.PermissionDenied, "can only copy your own personal virtual files")
		}
		sourceData = src.Data
		sourceDisplayName = srcDisplayName
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported source kind %q", req.SourceKind.String())
	}

	if displayName == "" {
		if sourceDisplayName != "" {
			displayName = "Copy of " + sourceDisplayName
		} else {
			displayName = "Copy of " + req.SourceName
		}
	}

	// Rewrite the YAML so it's owner-only and uses the new display name.
	copyData, err := rewritePersonalYAML(sourceData, displayName, ownerID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to prepare copy: %s", err.Error())
	}
	// Re-validate to ensure type & size invariants hold.
	copyData, err = validatePersonalYAML(spec, copyData, displayName, ownerID)
	if err != nil {
		return nil, err
	}

	name := randomPersonalName(displayName)
	err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
		ProjectID:   proj.ID,
		Environment: "prod",
		Path:        personalVirtualFilePath(spec, ownerID, name),
		Data:        copyData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert virtual file: %w", err)
	}

	err = s.admin.TriggerParserAndAwaitResource(ctx, depl, name, spec.runtimeKind)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile personal virtual file: %w", err)
	}

	return &adminv1.CopyPersonalVirtualFileResponse{Name: name}, nil
}

// GetPersonalVirtualFile returns the YAML body and metadata for the given personal virtual file.
// Only the owner can fetch.
func (s *Server) GetPersonalVirtualFile(ctx context.Context, req *adminv1.GetPersonalVirtualFileRequest) (*adminv1.GetPersonalVirtualFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.type", req.Type.String()),
		attribute.String("args.name", req.Name),
	)

	spec, ok := personalVirtualFileSpecs[req.Type]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal virtual file type %q", req.Type.String())
	}

	proj, _, claims, err := s.lookupProjectForPersonal(ctx, req.Org, req.Project, spec)
	if err != nil {
		return nil, err
	}

	vf, err := s.admin.DB.FindVirtualFile(ctx, proj.ID, "prod", personalVirtualFilePath(spec, claims.OwnerID(), req.Name))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "personal virtual file not found")
		}
		return nil, err
	}
	if vf.Deleted {
		return nil, status.Error(codes.NotFound, "personal virtual file not found")
	}

	ownerID, displayName, err := parsePersonalAnnotations(vf.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse file annotations: %s", err.Error())
	}
	if ownerID != claims.OwnerID() {
		return nil, status.Error(codes.NotFound, "personal virtual file not found")
	}

	return &adminv1.GetPersonalVirtualFileResponse{
		Name:        req.Name,
		DisplayName: displayName,
		Yaml:        string(vf.Data),
		UpdatedOn:   timestamppb.New(vf.UpdatedOn),
	}, nil
}

// ListPersonalVirtualFiles lists the caller's personal virtual files for the given type.
func (s *Server) ListPersonalVirtualFiles(ctx context.Context, req *adminv1.ListPersonalVirtualFilesRequest) (*adminv1.ListPersonalVirtualFilesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.type", req.Type.String()),
	)

	spec, ok := personalVirtualFileSpecs[req.Type]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported personal virtual file type %q", req.Type.String())
	}

	proj, _, claims, err := s.lookupProjectForPersonal(ctx, req.Org, req.Project, spec)
	if err != nil {
		return nil, err
	}

	prefix := personalVirtualFilePrefix(spec, claims.OwnerID())
	files, err := s.admin.DB.FindVirtualFilesByPrefix(ctx, proj.ID, "prod", prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual files: %w", err)
	}

	res := make([]*adminv1.PersonalVirtualFileSummary, 0, len(files))
	for _, vf := range files {
		name := personalNameFromPath(prefix, vf.Path)
		if name == "" {
			continue
		}
		_, displayName, err := parsePersonalAnnotations(vf.Data)
		if err != nil {
			// Skip unreadable rows rather than failing the whole listing.
			continue
		}
		res = append(res, &adminv1.PersonalVirtualFileSummary{
			Name:        name,
			DisplayName: displayName,
			Type:        req.Type,
			UpdatedOn:   timestamppb.New(vf.UpdatedOn),
		})
	}

	return &adminv1.ListPersonalVirtualFilesResponse{Files: res}, nil
}

// --- helpers ---

// lookupProjectForPersonal resolves the project, primary deployment, claims, and permission check
// shared by every personal virtual file handler.
func (s *Server) lookupProjectForPersonal(ctx context.Context, org, project string, spec *personalVirtualFileSpec) (*database.Project, *database.Deployment, auth.Claims, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, org, project)
	if err != nil {
		return nil, nil, nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProd {
		return nil, nil, nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}
	if !spec.hasPermission(permissions) {
		return nil, nil, nil, status.Error(codes.PermissionDenied, "does not have permission to manage personal canvases for this project")
	}
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, nil, nil, status.Error(codes.PermissionDenied, "only users can manage personal virtual files")
	}

	if proj.PrimaryDeploymentID == nil {
		return nil, nil, nil, status.Error(codes.FailedPrecondition, "project does not have a production deployment")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, nil, nil, err
	}

	return proj, depl, claims, nil
}

// fetchSharedSourceYAML reads a shared canvas (or other future type) from the runtime catalog and
// regenerates its YAML in a form suitable for use as a personal copy. Returns YAML bytes and the
// source's display name.
func (s *Server) fetchSharedSourceYAML(ctx context.Context, depl *database.Deployment, spec *personalVirtualFileSpec, sourceName, ownerID string) ([]byte, string, error) {
	switch spec.yamlType {
	case "canvas":
		canvasSpec, err := s.admin.LookupCanvas(ctx, depl, sourceName)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
				return nil, "", status.Errorf(codes.NotFound, "source canvas %q not found", sourceName)
			}
			return nil, "", fmt.Errorf("failed to look up source canvas: %w", err)
		}
		yamlBytes, err := yamlForManagedCanvas(canvasSpec.DisplayName, ownerID)
		if err != nil {
			return nil, "", status.Errorf(codes.Internal, "failed to generate copy YAML: %s", err.Error())
		}
		return yamlBytes, canvasSpec.DisplayName, nil
	default:
		return nil, "", status.Errorf(codes.Unimplemented, "copying from a shared %s is not supported yet", spec.yamlType)
	}
}

// validatePersonalYAML decodes the given YAML, asserts type == spec.yamlType, enforces the admin
// annotations (admin_managed=true, admin_owner_user_id, admin_nonce), and rewrites the document so
// the caller cannot smuggle conflicting annotations or steal someone else's ownership.
//
// If displayName is empty, the YAML's display_name is preserved; otherwise it is set to displayName.
func validatePersonalYAML(spec *personalVirtualFileSpec, data []byte, displayName, ownerID string) ([]byte, error) {
	if len(data) > maxPersonalYAMLSize {
		return nil, status.Errorf(codes.ResourceExhausted, "YAML body exceeds the %d byte limit", maxPersonalYAMLSize)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid YAML: %s", err.Error())
	}
	if doc == nil {
		doc = map[string]any{}
	}

	docType, _ := doc["type"].(string)
	if docType != spec.yamlType {
		return nil, status.Errorf(codes.InvalidArgument, "YAML type must be %q, got %q", spec.yamlType, docType)
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
	if len(out) > maxPersonalYAMLSize {
		return nil, status.Errorf(codes.ResourceExhausted, "YAML body exceeds the %d byte limit after normalization", maxPersonalYAMLSize)
	}
	return out, nil
}

// rewritePersonalYAML takes an existing YAML body (from a shared source or own copy) and rewrites
// annotations + display_name + security so the resulting body is owner-only and self-contained.
func rewritePersonalYAML(data []byte, displayName, ownerID string) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc == nil {
		doc = map[string]any{}
	}
	if displayName != "" {
		doc["display_name"] = displayName
	}
	// Strip any existing security stanza; the runtime built-in rule + admin_managed annotation
	// are sufficient and avoid surprising inherited policies.
	delete(doc, "security")
	annotations, _ := doc["annotations"].(map[string]any)
	if annotations == nil {
		annotations = map[string]any{}
	}
	annotations["admin_owner_user_id"] = ownerID
	annotations["admin_managed"] = true
	annotations["admin_nonce"] = time.Now().Format(time.RFC3339Nano)
	doc["annotations"] = annotations
	return yaml.Marshal(doc)
}

// parsePersonalAnnotations extracts admin_owner_user_id and display_name from a YAML body.
func parsePersonalAnnotations(data []byte) (ownerID, displayName string, err error) {
	var doc struct {
		DisplayName string            `yaml:"display_name"`
		Annotations map[string]string `yaml:"annotations"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", "", err
	}
	return doc.Annotations["admin_owner_user_id"], doc.DisplayName, nil
}

func personalVirtualFilePath(spec *personalVirtualFileSpec, ownerID, name string) string {
	return path.Join("personal", spec.pathPrefixSegment, ownerID, name+".yaml")
}

func personalVirtualFilePrefix(spec *personalVirtualFileSpec, ownerID string) string {
	return path.Join("personal", spec.pathPrefixSegment, ownerID) + "/"
}

// personalNameFromPath strips the prefix and ".yaml" suffix to recover the resource name.
// Returns "" if path does not have the expected shape.
func personalNameFromPath(prefix, fullPath string) string {
	if !strings.HasPrefix(fullPath, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(fullPath, prefix)
	if !strings.HasSuffix(rest, ".yaml") {
		return ""
	}
	return strings.TrimSuffix(rest, ".yaml")
}

// blankPersonalCanvasYAML returns the default YAML body for a new personal canvas.
func blankPersonalCanvasYAML(displayName, ownerID string) ([]byte, error) {
	return yamlForManagedCanvas(displayName, ownerID)
}

// yamlForManagedCanvas builds the YAML body for a personal canvas with owner-only annotations.
// Note: the actual access enforcement is the runtime's built-in admin-managed canvas rule; the
// security stanza below is a defence-in-depth marker that also makes the file self-documenting.
func yamlForManagedCanvas(displayName, ownerID string) ([]byte, error) {
	doc := map[string]any{
		"type":         "canvas",
		"display_name": displayName,
		"annotations": map[string]any{
			"admin_owner_user_id": ownerID,
			"admin_managed":       true,
			"admin_nonce":         time.Now().Format(time.RFC3339Nano),
		},
		"security": map[string]any{
			"access": fmt.Sprintf("'{{ .user.id }}' == '%s' || '{{ .user.admin }}' == 'true'", ownerID),
		},
		"rows": []any{},
	}
	return yaml.Marshal(doc)
}

var personalNameToDashCharsRegexp = regexp.MustCompile(`[ _]+`)

var personalNameExcludeCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// randomPersonalName returns a URL-safe slug derived from displayName with an 8-char UUID suffix
// (e.g. "My Canvas!" -> "my-canvas-5b3f7e1a"). The suffix makes collisions effectively impossible.
func randomPersonalName(displayName string) string {
	name := personalNameToDashCharsRegexp.ReplaceAllString(displayName, "-")
	name = personalNameExcludeCharsRegexp.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	name = strings.Trim(name, "-")
	if name == "" {
		return uuid.New().String()
	}
	return name + "-" + uuid.New().String()[0:8]
}

const maxPersonalYAMLSize = 128 * 1024 // matches the virtual_files DB limit
