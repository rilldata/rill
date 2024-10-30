package server

import (
	"context"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const gitURLTTL = 30 * time.Minute

func (s *Server) GetRepoMeta(ctx context.Context, req *adminv1.GetRepoMetaRequest) (*adminv1.GetRepoMetaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.String("args.branch", req.Branch),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProdStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ProdBranch != req.Branch {
		return nil, status.Error(codes.InvalidArgument, "branch not found")
	}

	if proj.ArchiveAssetID != nil {
		asset, err := s.admin.DB.FindAsset(ctx, *proj.ArchiveAssetID)
		if err != nil {
			return nil, err
		}

		downloadURL, err := s.generateV4GetObjectSignedURL(asset.Path)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return &adminv1.GetRepoMetaResponse{
			ArchiveDownloadUrl: downloadURL,
		}, nil
	}

	if proj.GithubURL == nil || proj.GithubInstallationID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a github integration")
	}

	token, err := s.admin.Github.InstallationToken(ctx, *proj.GithubInstallationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ep, err := transport.NewEndpoint(*proj.GithubURL + ".git") // TODO: Can the clone URL be different from the HTTP URL of a Github repo?
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create endpoint from %q: %s", *proj.GithubURL, err.Error())
	}
	ep.User = "x-access-token"
	ep.Password = token
	gitURL := ep.String()

	return &adminv1.GetRepoMetaResponse{
		GitUrl:          gitURL,
		GitUrlExpiresOn: timestamppb.New(time.Now().Add(gitURLTTL)),
		GitSubpath:      proj.Subpath,
	}, nil
}

func (s *Server) PullVirtualRepo(ctx context.Context, req *adminv1.PullVirtualRepoRequest) (*adminv1.PullVirtualRepoResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.String("args.branch", req.Branch),
		attribute.Int("args.page_size", int(req.PageSize)),
		attribute.String("args.page_token", req.PageToken),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	if proj.ProdBranch != req.Branch {
		return nil, status.Error(codes.InvalidArgument, "branch not found")
	}

	permissions := auth.GetClaims(ctx).ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProdStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	pageToken, err := unmarshalStringTimestampPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	vfs, err := s.admin.DB.FindVirtualFiles(ctx, proj.ID, req.Branch, pageToken.Ts.AsTime(), pageToken.Str, pageSize)
	if err != nil {
		return nil, err
	}

	// If no files were found, we return the same page token as the next page token.
	// This enables the client to poll for new changes continuously. (The client is responsible for pausing when an empty page is returned.)
	nextToken := req.PageToken
	if len(vfs) > 0 {
		f := vfs[len(vfs)-1]
		nextToken = marshalStringTimestampPageToken(f.Path, f.UpdatedOn)
	}

	dtos := make([]*adminv1.VirtualFile, len(vfs))
	for i, vf := range vfs {
		dtos[i] = virtualFileToDTO(vf)
	}

	return &adminv1.PullVirtualRepoResponse{
		Files:         dtos,
		NextPageToken: nextToken,
	}, nil
}

// objectpath is of form gs://<bucket>/.....
func (s *Server) generateV4GetObjectSignedURL(objectpath string) (string, error) {
	u, err := url.Parse(objectpath)
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

func virtualFileToDTO(vf *database.VirtualFile) *adminv1.VirtualFile {
	return &adminv1.VirtualFile{
		Path:      vf.Path,
		Data:      vf.Data,
		Deleted:   vf.Deleted,
		UpdatedOn: timestamppb.New(vf.UpdatedOn),
	}
}
