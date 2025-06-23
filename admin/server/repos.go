package server

import (
	"context"
	"time"

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

func (s *Server) GetRepoMeta(ctx context.Context, req *adminv1.GetRepoMetaRequest) (*adminv1.GetRepoMetaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	perms := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !perms.ReadProdStatus && !perms.ReadDevStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	if proj.ArchiveAssetID != nil {
		asset, err := s.admin.DB.FindAsset(ctx, *proj.ArchiveAssetID)
		if err != nil {
			return nil, err
		}

		downloadURL, err := s.generateSignedDownloadURL(asset)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return &adminv1.GetRepoMetaResponse{
			ValidUntilTime:     timestamppb.New(time.Now().Add(time.Hour * 24 * 365)), // Setting to a year because it doesn't need to be refreshed
			ArchiveId:          asset.ID,
			ArchiveDownloadUrl: downloadURL,
			ArchiveCreatedOn:   timestamppb.New(asset.CreatedOn),
		}, nil
	}

	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project does not have a github integration")
	}

	// nolint // Pending other PR merging.
	// var depl *database.Deployment
	// if claims.OwnerType() == auth.OwnerTypeDeployment {
	// 	var err error
	// 	depl, err = s.admin.DB.FindDeployment(ctx, claims.OwnerID())
	// 	if err != nil {
	// 		return nil, status.Error(codes.NotFound, "deployment not found")
	// 	}
	// }

	repoID, err := s.githubRepoIDForProject(ctx, proj)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.admin.Github.InstallationToken(ctx, *proj.GithubInstallationID, repoID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ep, err := transport.NewEndpoint(*proj.GitRemote)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create endpoint from %q: %s", *proj.GitRemote, err.Error())
	}
	ep.User = "x-access-token"
	ep.Password = token
	gitURL := ep.String()

	return &adminv1.GetRepoMetaResponse{
		ValidUntilTime: timestamppb.New(expiresAt),
		GitUrl:         gitURL,
		GitSubpath:     proj.Subpath,
		GitBranch:      proj.ProdBranch,
		// TODO: GitEditBranch from depl if not nil
	}, nil
}

func (s *Server) PullVirtualRepo(ctx context.Context, req *adminv1.PullVirtualRepoRequest) (*adminv1.PullVirtualRepoResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.ProjectId),
		attribute.Int("args.page_size", int(req.PageSize)),
		attribute.String("args.page_token", req.PageToken),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProdStatus && !permissions.ReadDevStatus {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project repo")
	}

	var depl *database.Deployment
	if claims.OwnerType() == auth.OwnerTypeDeployment {
		var err error
		depl, err = s.admin.DB.FindDeployment(ctx, claims.OwnerID())
		if err != nil {
			return nil, status.Error(codes.NotFound, "deployment not found")
		}
	}

	environment := "prod"
	if depl != nil { // nolint // Pending other PR merging.
		// TODO: Once deployments have environments.
		// environment = depl.Environment
	}

	pageToken, err := unmarshalStringTimestampPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	vfs, err := s.admin.DB.FindVirtualFiles(ctx, proj.ID, environment, pageToken.Ts.AsTime(), pageToken.Str, pageSize)
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

func virtualFileToDTO(vf *database.VirtualFile) *adminv1.VirtualFile {
	return &adminv1.VirtualFile{
		Path:      vf.Path,
		Data:      vf.Data,
		Deleted:   vf.Deleted,
		UpdatedOn: timestamppb.New(vf.UpdatedOn),
	}
}
