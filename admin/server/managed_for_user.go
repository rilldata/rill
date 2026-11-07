package server

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// managedFilesPageSize is the number of virtual files read at a time when listing a user's managed alerts or reports.
const managedFilesPageSize = 1000

// userForPrimaryDeployment checks that the caller is the project's primary deployment, and returns the user it acts for with their permissions on the project.
// The deployment is trusted to pass the user of its session, as for GetReportMeta and GetAlertMeta.
// Other deployments are rejected because managed alerts and reports live in the prod environment, which only the primary deployment reads.
// The user must be a user, not a service, magic token or embed user; what names the resources in errors, e.g. "alerts".
func (s *Server) userForPrimaryDeployment(ctx context.Context, proj *database.Project, userID, what string) (*database.User, *adminv1.ProjectPermissions, error) {
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeDeployment {
		return nil, nil, status.Errorf(codes.PermissionDenied, "only a deployment can act on %s for a user", what)
	}
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProject {
		return nil, nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}
	if proj.PrimaryDeploymentID == nil || *proj.PrimaryDeploymentID != claims.OwnerID() {
		return nil, nil, status.Errorf(codes.PermissionDenied, "only the project's primary deployment can act on %s for a user", what)
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, nil, status.Errorf(codes.PermissionDenied, "%s can only be owned by a user", what)
	}
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil, status.Errorf(codes.PermissionDenied, "%s can only be owned by a user", what)
		}
		return nil, nil, err
	}
	orgPerms, err := s.admin.OrganizationPermissionsForUser(ctx, proj.OrganizationID, user.ID)
	if err != nil {
		return nil, nil, err
	}
	projPerms, err := s.admin.ProjectPermissionsForUser(ctx, proj.ID, user.ID, orgPerms)
	if err != nil {
		return nil, nil, err
	}
	return user, projPerms, nil
}

// managedFilesForUser returns the virtual files of the managed alerts or reports (dir "alerts" or "reports") owned by a user in the prod environment.
func (s *Server) managedFilesForUser(ctx context.Context, projectID, dir, userID string) ([]*database.VirtualFile, error) {
	var res []*database.VirtualFile
	var afterUpdatedOn time.Time
	var afterPath string
	for {
		files, err := s.admin.DB.FindVirtualFiles(ctx, projectID, "prod", afterUpdatedOn, afterPath, managedFilesPageSize)
		if err != nil {
			return nil, err
		}
		for _, vf := range files {
			if vf.Deleted || !strings.HasPrefix(vf.Path, dir+"/") {
				continue
			}
			managed, owner := managedFileOwner(vf.Data)
			if managed && owner == userID {
				res = append(res, vf)
			}
		}
		if len(files) < managedFilesPageSize {
			return res, nil
		}
		last := files[len(files)-1]
		afterUpdatedOn, afterPath = last.UpdatedOn, last.Path
	}
}

// managedFileForEdit returns the virtual file of a managed alert or report the user owns, for an edit based on the given version of it.
// Unlike EditAlert and EditReport, only the owner can edit it: being a project admin doesn't make another user's alert or report theirs to change.
// It fails if the file changed since that version, so the edit doesn't overwrite a concurrent change.
func (s *Server) managedFileForEdit(ctx context.Context, projectID, path, kind, name, userID, version string) (*database.VirtualFile, error) {
	vf, err := s.admin.DB.FindVirtualFile(ctx, projectID, "prod", path)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "%s %q not found", kind, name)
		}
		return nil, err
	}
	if vf.Deleted {
		return nil, status.Errorf(codes.NotFound, "%s %q not found", kind, name)
	}
	if managed, owner := managedFileOwner(vf.Data); !managed || owner != userID {
		return nil, status.Errorf(codes.PermissionDenied, "does not have permission to edit %s", kind)
	}
	if version == "" {
		return nil, status.Errorf(codes.InvalidArgument, "version is required to edit %s %q", kind, name)
	}
	if version != managedFileVersion(vf.UpdatedOn) {
		return nil, managedFileChangedError(kind, name)
	}
	return vf, nil
}

// writeManagedFile writes the virtual file of a managed alert or report, and returns the version it wrote.
// If prev is not nil, it only overwrites that version of the file, and fails with codes.Aborted if it changed since.
func (s *Server) writeManagedFile(ctx context.Context, projectID, path string, data []byte, kind, name string, prev *database.VirtualFile) (string, error) {
	var err error
	if prev != nil {
		err = s.admin.DB.UpdateVirtualFileDataIfUnchanged(ctx, &database.UpdateVirtualFileDataOptions{
			ProjectID:   projectID,
			Environment: "prod",
			Path:        path,
			Data:        data,
			UpdatedOn:   prev.UpdatedOn,
		})
		if errors.Is(err, database.ErrNotFound) {
			return "", managedFileChangedError(kind, name)
		}
	} else {
		err = s.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
			ProjectID:    projectID,
			Environment:  "prod",
			Path:         path,
			Data:         data,
			OwnerID:      nil,
			EmbedOwnerID: nil,
		})
	}
	if err != nil {
		return "", err
	}

	// The writes don't return the version they wrote, so read it back.
	// The data carries a nonce, so if it differs, another write followed: then no version is returned, and an edit must list the file again.
	vf, err := s.admin.DB.FindVirtualFile(ctx, projectID, "prod", path)
	if err != nil {
		return "", err
	}
	if vf.Deleted || !bytes.Equal(vf.Data, data) {
		return "", nil
	}
	return managedFileVersion(vf.UpdatedOn), nil
}

// managedFileVersion identifies a version of a virtual file by when it was last updated.
func managedFileVersion(updatedOn time.Time) string {
	return updatedOn.UTC().Format(time.RFC3339Nano)
}

func managedFileChangedError(kind, name string) error {
	return status.Errorf(codes.Aborted, "%s %q changed since it was listed: list it again and reapply the change to its current version", kind, name)
}

// managedFileOwner returns whether a managed alert or report's YAML marks it as created from the UI, and its owner.
func managedFileOwner(data []byte) (bool, string) {
	var doc struct {
		Annotations struct {
			AdminManaged     bool   `yaml:"admin_managed"`
			AdminOwnerUserID string `yaml:"admin_owner_user_id"`
		} `yaml:"annotations"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, ""
	}
	return doc.Annotations.AdminManaged, doc.Annotations.AdminOwnerUserID
}

// resolverFromYAML returns the resolver and its properties from the data block of a managed alert or report, which has at most one key.
func resolverFromYAML(data map[string]any) (string, *structpb.Struct, error) {
	for resolver, v := range data {
		props, _ := v.(map[string]any)
		pb, err := structpb.NewStruct(props)
		if err != nil {
			return "", nil, err
		}
		return resolver, pb, nil
	}
	return "", nil, nil
}
