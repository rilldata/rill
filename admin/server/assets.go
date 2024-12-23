package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/archive"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// maxAssetSize is the maximum allowed size of a user-uploaded asset.
const maxAssetSize = 104857600 // 100 MB

// signingHeaders is a map of headers to be used when generating a signed URL for uploading an asset.
// It is used to enforce maxAssetSize.
var signingHeaders = map[string]string{
	"Content-Type":                "application/octet-stream",
	"x-goog-content-length-range": fmt.Sprintf("1,%d", maxAssetSize),
}

func (s *Server) CreateAsset(ctx context.Context, req *adminv1.CreateAssetRequest) (*adminv1.CreateAssetResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.OrganizationName),
		attribute.String("args.type", req.Type),
	)

	// Find the parent org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check permissions (create asset and create project should be the same permission)
	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).CreateProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create assets")
	}

	// Generate an ID and path for the asset
	assetID := uuid.New().String()
	objectPath := path.Join(req.Type, fmt.Sprintf("%s__%s__%s.%s", org.Name, req.Name, assetID, req.Extension))
	objectURL := &url.URL{
		Scheme: "gs",
		Host:   s.admin.Assets.BucketName(),
		Path:   objectPath,
	}

	// Generate a signed URL for uploading the asset
	var headers []string
	for k, v := range signingHeaders {
		headers = append(headers, fmt.Sprintf("%s:%s", k, v))
	}
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "PUT",
		Headers: headers,
		Expires: time.Now().Add(15 * time.Minute),
	}
	signedURL, err := s.admin.Assets.SignedURL(objectPath, opts)
	if err != nil {
		return nil, err
	}

	// Track the asset in the database.
	// If the upload fails or the asset is never linked to a use case, a background job will delete it after some time.
	asset, err := s.admin.DB.InsertAsset(ctx, assetID, org.ID, objectURL.String(), claims.OwnerID(), req.Cacheable)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert asset: %s", err.Error())
	}

	return &adminv1.CreateAssetResponse{
		AssetId:        asset.ID,
		SignedUrl:      signedURL,
		SigningHeaders: signingHeaders,
	}, nil
}

// assetHandler serves a previously uploaded file asset.
// If the asset is marked as cacheable, it sets caching headers that allows CDNs and browsers to cache the asset.
// If the asset is not marked as cacheable, it guarantees that the asset can only be accessed by authenticated users with read access to the asset's organization.
func (s *Server) assetHandler(w http.ResponseWriter, r *http.Request) error {
	// Params
	assetID := r.PathValue("asset_id")

	// Find the asset
	asset, err := s.admin.DB.FindAsset(r.Context(), assetID)
	if err != nil {
		return httputil.Error(http.StatusNotFound, err)
	}
	if asset.OrganizationID == nil {
		return httputil.Errorf(http.StatusNotFound, "the requested asset has been soft deleted")
	}

	// Check permissions
	claims := auth.GetClaims(r.Context())
	if !claims.OrganizationPermissions(r.Context(), *asset.OrganizationID).ReadOrg {
		return httputil.Errorf(http.StatusForbidden, "does not have permission to access the asset")
	}

	// Parse the asset's path, which has the form "gs://<bucket>/<path>"
	u, err := url.Parse(asset.Path)
	if err != nil {
		return err
	}

	// Download the asset and stream it to the client
	data, err := s.admin.Assets.Object(u.Path).NewReader(r.Context())
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return httputil.Error(http.StatusRequestTimeout, err)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}
	defer data.Close()

	// Set caching headers if the asset is cacheable
	if asset.Cacheable {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	} else {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(http.StatusOK)

	// Copy the data reader to the response writer
	_, err = io.Copy(w, data)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return httputil.Error(http.StatusRequestTimeout, err)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}

// generateSignedURL generates a signed URL for downloading the asset.
func (s *Server) generateSignedURL(asset *database.Asset) (string, error) {
	// asset.Path is of the form "gs://<bucket>/<path>"
	u, err := url.Parse(asset.Path)
	if err != nil {
		return "", err
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	}

	signedURL, err := s.admin.Assets.SignedURL(strings.TrimPrefix(u.Path, "/"), opts)
	if err != nil {
		return "", err
	}
	return signedURL, nil
}

// UploadProjectAssets disconnects a project from Github by uploading the contents of a Github repository as an asset.
func (s *Server) UploadProjectAssets(ctx context.Context, req *adminv1.UploadProjectAssetsRequest) (*adminv1.UploadProjectAssetsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Organization),
		attribute.String("args.project", req.Project),
	)

	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser && claims.OwnerType() != auth.OwnerTypeService {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Find parent org
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Organization, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check permissions
	// create asset and create project should be the same permission
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to edit project")
	}

	if proj.GithubURL == nil {
		return nil, status.Error(codes.InvalidArgument, "project is not connected to github")
	}

	assetResp, err := s.CreateAsset(ctx, &adminv1.CreateAssetRequest{
		OrganizationName: req.Organization,
		Type:             "deploy",
		Name:             req.Project,
		Extension:        "tar.gz",
	})
	if err != nil {
		return nil, err
	}

	token, err := s.admin.Github.InstallationToken(ctx, *proj.GithubInstallationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	archiveRoot, err := os.MkdirTemp(os.TempDir(), "archives")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(archiveRoot)

	files, err := gitToFilesList(archiveRoot, safeStr(proj.GithubURL), proj.ProdBranch, proj.Subpath, token)
	if err != nil {
		return nil, err
	}

	archivePath := archiveRoot
	if proj.Subpath != "" {
		archivePath = filepath.Join(archivePath, proj.Subpath)
	}
	err = archive.Create(ctx, files, archivePath, assetResp.SignedUrl, assetResp.SigningHeaders)
	if err != nil {
		return nil, err
	}

	_, err = s.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
		OrganizationName: req.Organization,
		Name:             req.Project,
		ArchiveAssetId:   &assetResp.AssetId,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.UploadProjectAssetsResponse{}, nil
}

func gitToFilesList(gitPath, repo, branch, subpath, token string) ([]drivers.DirEntry, error) {
	// projPath is actual path for project including any subpath within the git root
	projPath := gitPath
	if subpath != "" {
		projPath = filepath.Join(projPath, subpath)
	}
	err := os.MkdirAll(projPath, fs.ModePerm)
	if err != nil {
		return nil, err
	}

	_, err = git.PlainClone(gitPath, false, &git.CloneOptions{
		URL:           repo,
		Auth:          &githttp.BasicAuth{Username: "x-access-token", Password: token},
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone source git repo: %w", err)
	}

	srcProjDir := os.DirFS(projPath)
	var entries []drivers.DirEntry
	err = doublestar.GlobWalk(srcProjDir, "**", func(p string, d fs.DirEntry) error {
		// Ignore unnecessary paths
		if drivers.IsIgnored(p, nil) {
			return nil
		}

		entries = append(entries, drivers.DirEntry{
			Path:  p,
			IsDir: d.IsDir(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}
