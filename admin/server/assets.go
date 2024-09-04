package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/archive"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 100 MB
const maxAssetSize = 104857600

var signingHeaderMap = map[string]string{
	"Content-Type":                "application/octet-stream",
	"x-goog-content-length-range": fmt.Sprintf("1,%d", maxAssetSize),
}

// a copy of signingHeaderMap but kept in array form to pass to SignedURL API
var signingHeaders = []string{
	"Content-Type:application/octet-stream",
	fmt.Sprintf("x-goog-content-length-range:1,%d", maxAssetSize), // validates that the request body is between 1 byte to 100MB
}

func (s *Server) CreateAsset(ctx context.Context, req *adminv1.CreateAssetRequest) (*adminv1.CreateAssetResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.OrganizationName),
		attribute.String("args.type", req.Type),
	)

	// Check the request is made by a user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser && claims.OwnerType() != auth.OwnerTypeService {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Find parent org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.OrganizationName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check permissions
	// create asset and create project should be the same permission
	if !claims.OrganizationPermissions(ctx, org.ID).CreateProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create assets")
	}

	// generate a signed url
	object := path.Join(req.Type, fmt.Sprintf("%s__%s.%s", req.Name, uuid.New().String(), req.Extension))
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "PUT",
		Headers: signingHeaders,
		Expires: time.Now().Add(15 * time.Minute),
	}
	signedURL, err := s.admin.Assets.SignedURL(object, opts)
	if err != nil {
		return nil, err
	}

	// create an asset
	assetPath, err := s.assetPath(object)
	if err != nil {
		return nil, err
	}

	asset, err := s.admin.DB.InsertAsset(ctx, org.ID, assetPath, claims.OwnerID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to insert asset: %s", err.Error())
	}

	return &adminv1.CreateAssetResponse{
		AssetId:        asset.ID,
		SignedUrl:      signedURL,
		SigningHeaders: signingHeaderMap,
	}, nil
}

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
		Name:             fmt.Sprintf("%s__%s", req.Organization, req.Project),
		Extension:        "tar.gz",
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
		OrganizationName: req.Organization,
		Name:             req.Project,
		ArchiveAssetId:   &assetResp.AssetId,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.UploadProjectAssetsResponse{}, nil
}

func (s *Server) assetPath(object string) (string, error) {
	uploadPath, err := url.Parse(s.opts.AssetsBucket)
	if err != nil {
		return "", err
	}
	uploadPath.Host = s.opts.AssetsBucket
	uploadPath.Scheme = "gs"
	uploadPath.Path = object
	return uploadPath.String(), nil
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
