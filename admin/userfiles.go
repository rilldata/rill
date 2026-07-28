package admin

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/rilldata/rill/admin/database"
)

// User files are UI-created resources (alerts, reports, personal canvases) that are staged in the
// virtual_files table and promoted into the project's Git repo under:
//
//	user_files/<segment>/<subdir>/<name>.yaml
//
// The <segment> identifies the owning user in a human-readable but collision-free way. It is purely
// cosmetic for lookups: edits and deletes find the row by resource name (see FindVirtualFileByName),
// so a later change to the user's name or email does not orphan their files.
const (
	UserFilesRoot          = "user_files"
	UserFilesSharedSegment = "_shared"
	UserFilesSubdirAlert   = "alerts"
	UserFilesSubdirReport  = "reports"
	UserFilesSubdirCanvas  = "canvas"
)

var (
	userFilesSegmentToDashCharsRegexp = regexp.MustCompile(`[ _.]+`)
	userFilesSegmentExcludeRegexp     = regexp.MustCompile(`[^a-zA-Z0-9-]+`)
)

// UserFilesDirSegment returns the directory segment for a user: "<name>-<userID>",
// e.g. "jane-doe-018f3a2b-7c4d-4e5f-9a1b-2c3d4e5f6a7b".
// The human-readable part comes from the user's display name, falling back to the email local part.
// The full user ID makes the segment unique. An empty userID (e.g. a service account) maps to the
// shared segment.
func UserFilesDirSegment(userID, displayName, email string) string {
	if userID == "" {
		return UserFilesSharedSegment
	}

	name := sanitizeUserFilesSegmentPart(displayName)
	if name == "" {
		local := email
		if i := strings.IndexByte(local, '@'); i >= 0 {
			local = local[:i]
		}
		name = sanitizeUserFilesSegmentPart(local)
	}
	if name == "" {
		name = "user"
	}
	return name + "-" + userID
}

func sanitizeUserFilesSegmentPart(s string) string {
	s = userFilesSegmentToDashCharsRegexp.ReplaceAllString(s, "-")
	s = userFilesSegmentExcludeRegexp.ReplaceAllString(s, "")
	return strings.ToLower(strings.Trim(s, "-"))
}

// UserFilesPath builds the repo-relative path for a user file.
func UserFilesPath(segment, subdir, name string) string {
	return path.Join(UserFilesRoot, segment, subdir, name+".yaml")
}

// UserFilesPathForOwner resolves the owning user and returns the repo-relative path for a new user file.
// ownerUserID may be empty or belong to a non-user owner (e.g. a service), in which case the shared
// segment is used.
func (s *Service) UserFilesPathForOwner(ctx context.Context, ownerUserID, subdir, name string) (string, error) {
	if ownerUserID == "" {
		return UserFilesPath(UserFilesSharedSegment, subdir, name), nil
	}
	user, err := s.DB.FindUser(ctx, ownerUserID)
	if err != nil {
		// Non-user owners (e.g. service tokens) aren't in the users table; their files go to the shared segment.
		if errors.Is(err, database.ErrNotFound) {
			return UserFilesPath(UserFilesSharedSegment, subdir, name), nil
		}
		return "", err
	}
	return UserFilesPath(UserFilesDirSegment(user.ID, user.DisplayName, user.Email), subdir, name), nil
}

// UserFilesRewriteLegacyPath returns the user_files subdir for a virtual file stored at a legacy path
// ("alerts/", "reports/", or "personal/" at the repo root), along with the resource name.
// It returns ok=false for paths already in the user_files layout (or any other path).
func UserFilesRewriteLegacyPath(p string) (subdir, name string, ok bool) {
	dir, file, found := strings.Cut(p, "/")
	if !found || !strings.HasSuffix(file, ".yaml") || strings.Contains(file, "/") {
		return "", "", false
	}
	name = strings.TrimSuffix(file, ".yaml")
	switch dir {
	case UserFilesSubdirAlert:
		return UserFilesSubdirAlert, name, true
	case UserFilesSubdirReport:
		return UserFilesSubdirReport, name, true
	case "personal": // Personal files were stored under "personal/" before the user_files layout.
		return UserFilesSubdirCanvas, name, true
	default:
		return "", "", false
	}
}

// UserFilesPathForVirtualFile returns the target user_files path for an existing virtual file row,
// rewriting legacy paths into the new layout. It returns ok=false if the row is already in the new
// layout (or at an unrecognized path).
func (s *Service) UserFilesPathForVirtualFile(ctx context.Context, vf *database.VirtualFile) (string, bool, error) {
	subdir, name, ok := UserFilesRewriteLegacyPath(vf.Path)
	if !ok {
		return "", false, nil
	}
	var ownerID string
	if vf.OwnerID != nil {
		ownerID = *vf.OwnerID
	}
	newPath, err := s.UserFilesPathForOwner(ctx, ownerID, subdir, name)
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve user files path for %q: %w", vf.Path, err)
	}
	return newPath, true, nil
}
