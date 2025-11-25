package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/rilldata/rill/admin/pkg/publicemail"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/env"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const devDeplTTL = 6 * time.Hour

const devSlots = 8

const prodDeplTTL = 14 * 24 * time.Hour

// runtimeAccessTokenTTL is the validity duration of JWTs issued for runtime access when calling GetProject.
// This TTL is not used for tokens created for internal communication between the admin and runtime services.
const runtimeAccessTokenDefaultTTL = 30 * time.Minute

// runtimeAccessTokenEmbedTTL is the validation duration of JWTs issued for embedding.
// Since low-risk embed users might not implement refresh, it defaults to a high value of 24 hours.
// It can be overridden to a lower value when issued for high-risk embed users.
const runtimeAccessTokenEmbedTTL = 24 * time.Hour

func (s *Server) ListProjectsForOrganization(ctx context.Context, req *adminv1.ListProjectsForOrganizationRequest) (*adminv1.ListProjectsForOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	// If user has ManageProjects, return all projects
	claims := auth.GetClaims(ctx)
	var projs []*database.Project
	if claims.OrganizationPermissions(ctx, org.ID).ManageProjects {
		projs, err = s.admin.DB.FindProjectsForOrganization(ctx, org.ID, token.Val, pageSize)
	} else if claims.OwnerType() == auth.OwnerTypeUser {
		// Get projects the user is a (direct or group) member of, plus all public projects.
		projs, err = s.admin.DB.FindProjectsForOrgAndUser(ctx, org.ID, claims.OwnerID(), true, token.Val, pageSize)
	} else {
		projs, err = s.admin.DB.FindPublicProjectsInOrganization(ctx, org.ID, token.Val, pageSize)
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// If no projects are public, and user is not an outside member of any projects, the projsMap is empty.
	// If additionally, the user is not an org member, return permission denied (instead of an empty slice).
	if len(projs) == 0 && !claims.OrganizationPermissions(ctx, org.ID).ReadProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read projects")
	}

	nextToken := ""
	if len(projs) >= pageSize {
		nextToken = marshalPageToken(projs[len(projs)-1].Name)
	}

	dtos := make([]*adminv1.Project, len(projs))
	for i, p := range projs {
		dtos[i] = s.projToDTO(p, org.Name)
	}

	return &adminv1.ListProjectsForOrganizationResponse{
		Projects:      dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListProjectsForOrganizationAndUser(ctx context.Context, req *adminv1.ListProjectsForOrganizationAndUserRequest) (*adminv1.ListProjectsForOrganizationAndUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.user_id", req.UserId),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read org members")
	}

	pageToken, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	projects, err := s.admin.DB.FindProjectsForOrgAndUser(ctx, org.ID, req.UserId, false, pageToken.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(projects) >= pageSize {
		nextToken = marshalPageToken(projects[len(projects)-1].Name)
	}

	dtos := make([]*adminv1.Project, len(projects))
	for i, p := range projects {
		dtos[i] = s.projToDTO(p, org.Name)
	}

	return &adminv1.ListProjectsForOrganizationAndUserResponse{
		Projects:      dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListProjectsForFingerprint(ctx context.Context, req *adminv1.ListProjectsForFingerprintRequest) (*adminv1.ListProjectsForFingerprintResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.directory_name", req.DirectoryName),
		attribute.String("args.git_remote", req.GitRemote),
		attribute.String("args.sub_path", req.SubPath),
		attribute.String("args.rill_mgd_git_remote", req.RillMgdGitRemote),
	)

	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can list projects by fingerprint")
	}
	userID := claims.OwnerID()

	// check if rill mgd remote was transferred
	// we do not support transfers from self hosted git repos so no need to check for that
	rillMgdRemote := req.RillMgdGitRemote
	transfer, err := s.admin.DB.FindGitRepoTransfer(ctx, rillMgdRemote)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if transfer != nil {
		rillMgdRemote = transfer.To
	}

	projects, err := s.admin.DB.FindProjectsForUserAndFingerprint(ctx, userID, req.DirectoryName, normalizeGitRemote(req.GitRemote), req.SubPath, rillMgdRemote)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 && req.GitRemote != "" {
		// if no project is found check if there is project user doesn't have access to
		projects, err = s.admin.DB.FindProjectsByGitRemote(ctx, normalizeGitRemote(req.GitRemote))
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			if p.Subpath != req.SubPath {
				continue
			}
			org, err := s.admin.DB.FindOrganization(ctx, p.OrganizationID)
			if err != nil {
				return nil, err
			}
			return &adminv1.ListProjectsForFingerprintResponse{
				UnauthorizedProject: fmt.Sprintf("%s/%s", org.Name, p.Name),
			}, nil
		}
		return &adminv1.ListProjectsForFingerprintResponse{}, nil
	}

	dtos := make([]*adminv1.Project, len(projects))
	orgNames := make(map[string]string)
	for i, p := range projects {
		orgName := orgNames[p.OrganizationID]
		if orgName == "" {
			org, err := s.admin.DB.FindOrganization(ctx, p.OrganizationID)
			if err != nil {
				return nil, err
			}
			orgName = org.Name
			orgNames[p.OrganizationID] = orgName
		}

		dtos[i] = s.projToDTO(p, orgName)
	}

	return &adminv1.ListProjectsForFingerprintResponse{
		Projects: dtos,
	}, nil
}

func (s *Server) ListProjectsForUserByName(ctx context.Context, req *adminv1.ListProjectsForUserByNameRequest) (*adminv1.ListProjectsForUserByNameResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project", req.Name),
	)

	claims := auth.GetClaims(ctx)
	userID := claims.OwnerID()

	projects, err := s.admin.DB.FindProjectsByNameAndUser(ctx, req.Name, userID)
	if err != nil {
		return nil, err
	}

	orgsByID := make(map[string]*database.Organization)

	dtos := make([]*adminv1.Project, len(projects))
	for i, p := range projects {
		org, hasOrg := orgsByID[p.OrganizationID]
		if !hasOrg {
			org, err = s.admin.DB.FindOrganization(ctx, p.OrganizationID)
			if err != nil {
				return nil, err
			}
			orgsByID[p.OrganizationID] = org
		}

		dtos[i] = s.projToDTO(p, org.Name)
	}

	return &adminv1.ListProjectsForUserByNameResponse{
		Projects: dtos,
	}, nil
}

func (s *Server) GetProject(ctx context.Context, req *adminv1.GetProjectRequest) (*adminv1.GetProjectResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}
	if forceAccess {
		permissions.ReadProject = true
		permissions.ReadProd = true
		permissions.ReadProdStatus = true
		permissions.ReadDev = true
		permissions.ReadDevStatus = true
		permissions.ReadProvisionerResources = true
		permissions.ReadProjectMembers = true
	}

	if !permissions.ReadProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	if proj.ProdDeploymentID == nil || !permissions.ReadProd {
		return &adminv1.GetProjectResponse{
			Project:            s.projToDTO(proj, org.Name),
			ProjectPermissions: permissions,
		}, nil
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !permissions.ReadProdStatus {
		depl.StatusMessage = ""
	}

	var attr map[string]any
	var rules []*runtimev1.SecurityRule
	if claims.OwnerType() == auth.OwnerTypeUser {
		attr, err = s.jwtAttributesForUser(ctx, claims.OwnerID(), proj.OrganizationID, permissions)
		if err != nil {
			return nil, err
		}
	} else if claims.OwnerType() == auth.OwnerTypeService {
		attr, err = s.jwtAttributesForService(ctx, claims.OwnerID(), permissions)
		if err != nil {
			return nil, err
		}
	} else if claims.OwnerType() == auth.OwnerTypeMagicAuthToken {
		mdl, ok := claims.AuthTokenModel().(*database.MagicAuthToken)
		if !ok {
			return nil, status.Errorf(codes.Internal, "unexpected type %T for magic auth token model", claims.AuthTokenModel())
		}

		for _, r := range mdl.Resources {
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{
							Kind: r.Type,
							Name: r.Name,
						},
					},
				},
			})
		}

		attr = mdl.Attributes
		if mdl.FilterJSON != "" {
			expr := &runtimev1.Expression{}
			err := protojson.Unmarshal([]byte(mdl.FilterJSON), expr)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "could not unmarshal metrics view filter: %s", err.Error())
			}

			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_RowFilter{
					RowFilter: &runtimev1.SecurityRuleRowFilter{
						Expression: expr,
					},
				},
			})
		}

		if len(mdl.Fields) > 0 {
			rules = append(rules, &runtimev1.SecurityRule{
				Rule: &runtimev1.SecurityRule_FieldAccess{
					FieldAccess: &runtimev1.SecurityRuleFieldAccess{
						Fields:    mdl.Fields,
						Allow:     true,
						Exclusive: true,
					},
				},
			})
		}
	}

	ttlDuration := runtimeAccessTokenDefaultTTL
	if req.AccessTokenTtlSeconds != 0 {
		ttlDuration = time.Duration(req.AccessTokenTtlSeconds) * time.Second
	}

	instancePermissions := []runtime.Permission{
		runtime.ReadObjects,
		runtime.ReadMetrics,
		runtime.ReadAPI,
		runtime.UseAI,
	}
	if permissions.ManageProject {
		instancePermissions = append(instancePermissions, runtime.EditTrigger, runtime.ReadResolvers)
	}

	var systemPermissions []runtime.Permission
	if req.IssueSuperuserToken {
		if !claims.Superuser(ctx) {
			return nil, status.Error(codes.PermissionDenied, "only superusers can issue superuser tokens")
		}
		// NOTE: The ManageInstances permission is currently used by the runtime to skip access checks.
		systemPermissions = append(systemPermissions, runtime.ManageInstances)
	}

	jwt, err := s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL:       depl.RuntimeAudience,
		Subject:           claims.OwnerID(),
		TTL:               ttlDuration,
		SystemPermissions: systemPermissions,
		InstancePermissions: map[string][]runtime.Permission{
			depl.RuntimeInstanceID: instancePermissions,
		},
		Attributes:    attr,
		SecurityRules: rules,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not issue jwt: %s", err.Error())
	}

	s.admin.Used.Deployment(depl.ID)

	return &adminv1.GetProjectResponse{
		Project:            s.projToDTO(proj, org.Name),
		ProdDeployment:     deploymentToDTO(depl),
		Jwt:                jwt,
		ProjectPermissions: permissions,
	}, nil
}

func (s *Server) GetProjectByID(ctx context.Context, req *adminv1.GetProjectByIDRequest) (*adminv1.GetProjectByIDResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.project_id", req.Id),
	)

	proj, err := s.admin.DB.FindProject(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if !permissions.ReadProject && !proj.Public && !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project")
	}

	return &adminv1.GetProjectByIDResponse{
		Project: s.projToDTO(proj, org.Name),
	}, nil
}

func (s *Server) SearchProjectNames(ctx context.Context, req *adminv1.SearchProjectNamesRequest) (*adminv1.SearchProjectNamesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.pattern", req.NamePattern),
		attribute.Int("args.annotations", len(req.Annotations)),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can search projects")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	var projectNames []string
	if len(req.Annotations) > 0 {
		// If an annotation is set to "*", we just check for key presence (instead of exact key-value match)
		var annotationKeys []string
		for k, v := range req.Annotations {
			if v == "*" {
				annotationKeys = append(annotationKeys, k)
				delete(req.Annotations, k)
			}
		}

		projectNames, err = s.admin.DB.FindProjectPathsByPatternAndAnnotations(ctx, req.NamePattern, token.Val, annotationKeys, req.Annotations, pageSize)
	} else {
		projectNames, err = s.admin.DB.FindProjectPathsByPattern(ctx, req.NamePattern, token.Val, pageSize)
	}
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(projectNames) >= pageSize {
		nextToken = marshalPageToken(projectNames[len(projectNames)-1])
	}

	return &adminv1.SearchProjectNamesResponse{
		Names:         projectNames,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) CreateProject(ctx context.Context, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.description", req.Description),
		attribute.Bool("args.public", req.Public),
		attribute.String("args.directory_name", req.DirectoryName),
		attribute.String("args.provisioner", req.Provisioner),
		attribute.String("args.prod_version", req.ProdVersion),
		attribute.Int64("args.prod_slots", req.ProdSlots),
		attribute.String("args.sub_path", req.Subpath),
		attribute.String("args.prod_branch", req.ProdBranch),
		attribute.String("args.git_remote", req.GitRemote),
		attribute.String("args.archive_asset_id", req.ArchiveAssetId),
		attribute.Bool("args.skip_deploy", req.SkipDeploy),
	)

	// Backwards compatibility
	req.GitRemote = normalizeGitRemote(req.GitRemote)

	// Find parent org
	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check permissions
	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).CreateProjects {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to create projects")
	}

	// check if org has any blocking billing errors
	err = s.admin.CheckBlockingBillingErrors(ctx, org.ID)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			return nil, err
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check projects quota
	usage, err := s.admin.DB.CountProjectsQuotaUsage(ctx, org.ID)
	if err != nil {
		return nil, err
	}
	if org.QuotaProjects >= 0 && usage.Projects >= org.QuotaProjects {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d projects", org.Name, org.QuotaProjects)
	}
	if org.QuotaSlotsPerDeployment >= 0 && int(req.ProdSlots) > org.QuotaSlotsPerDeployment {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org can't provision more than %d slots per deployment", org.QuotaSlotsPerDeployment)
	}
	if org.QuotaSlotsTotal >= 0 && usage.Slots+int(req.ProdSlots) > org.QuotaSlotsTotal {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d total slots", org.Name, org.QuotaSlotsTotal)
	}
	if org.QuotaDeployments >= 0 && usage.Deployments >= org.QuotaDeployments {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q is limited to %d deployments", org.Name, org.QuotaDeployments)
	}

	// Add prod TTL as 14 days if not a public project else infinite
	var prodTTL *int64
	if !req.Public {
		tmp := int64(prodDeplTTL.Seconds())
		prodTTL = &tmp
	}

	// Add dev TTL as 6 hours
	devTTL := int64(devDeplTTL.Seconds())

	// Backwards compatibility: if prod version is not set, default to "latest"
	if req.ProdVersion == "" {
		req.ProdVersion = "latest"
	}

	// Capture creating user (can be nil if created with a service token)
	var userID *string
	if claims.OwnerType() == auth.OwnerTypeUser {
		tmp := claims.OwnerID()
		userID = &tmp
	}

	// Prepare the project options
	opts := &database.InsertProjectOptions{
		OrganizationID:       org.ID,
		Name:                 req.Project,
		Description:          req.Description,
		Public:               req.Public,
		CreatedByUserID:      userID,
		DirectoryName:        req.DirectoryName,
		Provisioner:          req.Provisioner,
		ArchiveAssetID:       nil,         // Populated below
		GitRemote:            nil,         // Populated below
		GithubInstallationID: nil,         // Populated below
		GithubRepoID:         nil,         // Populated below
		ManagedGitRepoID:     nil,         // Populated below
		ProdBranch:           "",          // Populated below
		Subpath:              req.Subpath, // Populated below
		ProdVersion:          req.ProdVersion,
		ProdSlots:            int(req.ProdSlots),
		ProdTTLSeconds:       prodTTL,
		DevSlots:             devSlots,
		DevTTLSeconds:        devTTL,
	}

	// Check and validate the project file source.
	// NOTE: It is allowed to create a project without a source. It will then error later when creating the deployment (which can be skipped by passing skip_deploy).
	if req.GitRemote != "" && req.ArchiveAssetId != "" {
		return nil, status.Error(codes.InvalidArgument, "cannot set both git_remote and archive_asset_id")
	} else if req.GitRemote != "" {
		opts.GithubRepoID, opts.GithubInstallationID, opts.ManagedGitRepoID, opts.ProdBranch, err = s.githubOptsForRemote(ctx, org.ID, req.ProdBranch, userID, req.GitRemote)
		if err != nil {
			return nil, err
		}
		opts.GitRemote = &req.GitRemote
		opts.Subpath = req.Subpath
	} else if req.ArchiveAssetId != "" {
		// Check access to the archive asset
		if !s.hasAssetUsagePermission(ctx, req.ArchiveAssetId, org.ID, claims.OwnerID()) {
			return nil, status.Error(codes.PermissionDenied, "archive_asset_id is not accessible to this org")
		}
		opts.ArchiveAssetID = &req.ArchiveAssetId
	}

	// if there is no subscription for the org, submit a job to start a trial
	bi, err := s.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeNeverSubscribed)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if bi != nil {
		// check against trial orgs quota but skip if the user is a superuser
		if org.CreatedByUserID != nil && !claims.Superuser(ctx) {
			u, err := s.admin.DB.FindUser(ctx, *org.CreatedByUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to find user: %w", err)
			}
			if u.QuotaTrialOrgs >= 0 && u.CurrentTrialOrgsCount >= u.QuotaTrialOrgs {
				return nil, status.Errorf(codes.FailedPrecondition, "trial orgs quota exceeded for user %s", u.Email)
			}
		}
		_, err = s.admin.Jobs.StartOrgTrial(ctx, org.ID)
		if err != nil {
			s.logger.Named("billing").Error("failed to submit job to start trial for org, please do it manually", zap.String("org_id", org.ID), zap.Error(err))
			// continue creating the project
		}
	}

	// Create the project
	proj, err := s.admin.CreateProject(ctx, org, opts, !req.SkipDeploy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateProjectResponse{
		Project: s.projToDTO(proj, org.Name),
	}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *adminv1.DeleteProjectRequest) (*adminv1.DeleteProjectResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to delete project")
	}

	err = s.admin.TeardownProject(ctx, proj)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.DeleteProjectResponse{
		Id: proj.ID,
	}, nil
}

func (s *Server) UpdateProject(ctx context.Context, req *adminv1.UpdateProjectRequest) (*adminv1.UpdateProjectResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)
	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}
	if req.Description != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.description", *req.Description))
	}
	if req.Public != nil {
		observability.AddRequestAttributes(ctx, attribute.Bool("args.public", *req.Public))
	}
	if req.DirectoryName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.directory_name", *req.DirectoryName))
	}
	if req.Provisioner != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.provisioner", *req.Provisioner))
	}
	if req.ProdVersion != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.prod_version", *req.ProdVersion))
	}
	if req.ProdBranch != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.prod_branch", *req.ProdBranch))
	}
	if req.GitRemote != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.git_remote", *req.GitRemote))
	}
	if req.Subpath != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.subpath", *req.Subpath))
	}
	if req.ArchiveAssetId != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.archive_asset_id", *req.ArchiveAssetId))
	}
	if req.Public != nil {
		observability.AddRequestAttributes(ctx, attribute.Bool("args.public", *req.Public))
	}
	if req.ProdSlots != nil {
		observability.AddRequestAttributes(ctx, attribute.Int64("args.prod_slots", *req.ProdSlots))
	}
	if req.ProdTtlSeconds != nil {
		observability.AddRequestAttributes(ctx, attribute.Int64("args.prod_ttl_seconds", *req.ProdTtlSeconds))
	}
	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}

	// Backwards compatibility
	if req.GitRemote != nil {
		*req.GitRemote = normalizeGitRemote(*req.GitRemote)
	}

	// Find project
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage project")
	}

	if req.GitRemote != nil && req.ArchiveAssetId != nil {
		return nil, fmt.Errorf("cannot set both git_remote and archive_asset_id")
	}
	gitRemote := proj.GitRemote
	githubInstID := proj.GithubInstallationID
	githubRepoID := proj.GithubRepoID
	managedGitRepoID := proj.ManagedGitRepoID
	subpath := valOrDefault(req.Subpath, proj.Subpath)
	prodBranch := valOrDefault(req.ProdBranch, proj.ProdBranch)
	archiveAssetID := proj.ArchiveAssetID

	transferRepo := false
	var oldRemote string
	if req.GitRemote != nil && safeStr(proj.GitRemote) != *req.GitRemote {
		// check if another project deploys using the same git remote + subpath
		projects, err := s.admin.DB.FindProjectsByGitRemote(ctx, *req.GitRemote)
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			if p.ID == proj.ID {
				continue
			}
			if p.Subpath == subpath {
				org, err := s.admin.DB.FindOrganization(ctx, p.OrganizationID)
				if err != nil {
					return nil, err
				}
				return nil, status.Errorf(codes.FailedPrecondition, "another project %q in org %q is already using the same git remote and subpath", p.Name, org.Name)
			}
		}

		// check the Github app is installed and caller has access on the repo
		var userID *string
		if claims.OwnerType() == auth.OwnerTypeUser {
			tmp := claims.OwnerID()
			userID = &tmp
		}
		githubRepoID, githubInstID, managedGitRepoID, prodBranch, err = s.githubOptsForRemote(ctx, proj.OrganizationID, prodBranch, userID, *req.GitRemote)
		if err != nil {
			return nil, err
		}
		if managedGitRepoID != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid git remote: cannot switch to a rill managed git repo")
		}

		gitRemote = req.GitRemote
		managedGitRepoID = nil
		archiveAssetID = nil
		if proj.ManagedGitRepoID != nil {
			transferRepo = true
			oldRemote = *proj.GitRemote
		}
	}
	if req.ArchiveAssetId != nil {
		archiveAssetID = req.ArchiveAssetId
		org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
		if err != nil {
			return nil, err
		}
		if !s.hasAssetUsagePermission(ctx, *archiveAssetID, org.ID, claims.OwnerID()) {
			return nil, status.Error(codes.PermissionDenied, "archive_asset_id is not accessible to this org")
		}
		gitRemote = nil
		githubInstID = nil
		subpath = ""
		prodBranch = ""
	}

	prodTTLSeconds := proj.ProdTTLSeconds
	if req.ProdTtlSeconds != nil {
		if *req.ProdTtlSeconds == 0 {
			prodTTLSeconds = nil
		} else {
			prodTTLSeconds = req.ProdTtlSeconds
		}
	}

	opts := &database.UpdateProjectOptions{
		Name:                 valOrDefault(req.NewName, proj.Name),
		Description:          valOrDefault(req.Description, proj.Description),
		Public:               valOrDefault(req.Public, proj.Public),
		DirectoryName:        valOrDefault(req.DirectoryName, proj.DirectoryName),
		ArchiveAssetID:       archiveAssetID,
		GitRemote:            gitRemote,
		GithubInstallationID: githubInstID,
		GithubRepoID:         githubRepoID,
		ManagedGitRepoID:     managedGitRepoID,
		Subpath:              subpath,
		ProdVersion:          valOrDefault(req.ProdVersion, proj.ProdVersion),
		ProdBranch:           prodBranch,
		ProdDeploymentID:     proj.ProdDeploymentID,
		ProdSlots:            int(valOrDefault(req.ProdSlots, int64(proj.ProdSlots))),
		ProdTTLSeconds:       prodTTLSeconds,
		DevSlots:             proj.DevSlots,
		DevTTLSeconds:        proj.DevTTLSeconds,
		Provisioner:          valOrDefault(req.Provisioner, proj.Provisioner),
		Annotations:          proj.Annotations,
	}
	proj, err = s.admin.UpdateProject(ctx, proj, opts)
	if err != nil {
		return nil, err
	}

	// mark transfer from rill managed git repo if applicable
	if transferRepo {
		_, err = s.admin.DB.InsertGitRepoTransfer(ctx, oldRemote, *proj.GitRemote)
		if err != nil {
			return nil, err
		}
	}

	return &adminv1.UpdateProjectResponse{
		Project: s.projToDTO(proj, req.Org),
	}, nil
}

func (s *Server) GetProjectVariables(ctx context.Context, req *adminv1.GetProjectVariablesRequest) (*adminv1.GetProjectVariablesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.environment", req.Environment),
		attribute.Bool("args.for_all_environments", req.ForAllEnvironments),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to read project variables")
	}

	var vars []*database.ProjectVariable
	if req.ForAllEnvironments {
		vars, err = s.admin.DB.FindProjectVariables(ctx, proj.ID, nil)
	} else {
		vars, err = s.admin.DB.FindProjectVariables(ctx, proj.ID, &req.Environment)
	}
	if err != nil {
		return nil, err
	}

	resp := &adminv1.GetProjectVariablesResponse{
		Variables:    make([]*adminv1.ProjectVariable, 0, len(vars)),
		VariablesMap: make(map[string]string, len(vars)),
	}
	for _, v := range vars {
		resp.Variables = append(resp.Variables, projectVariableToDTO(v))
		// nolint:staticcheck // We still need to set it
		resp.VariablesMap[v.Name] = v.Value
	}
	return resp, nil
}

func (s *Server) UpdateProjectVariables(ctx context.Context, req *adminv1.UpdateProjectVariablesRequest) (*adminv1.UpdateProjectVariablesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.environment", req.Environment),
		attribute.StringSlice("args.variables", maps.Keys(req.Variables)),
		attribute.StringSlice("args.unset_variables", req.UnsetVariables),
	)
	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.PermissionDenied, "only users can update project variables")
	}
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to update project variables")
	}

	var validationErr error
	for k := range req.Variables {
		if err := env.ValidateName(k); err != nil {
			validationErr = errors.Join(validationErr, err)
		}
	}
	if validationErr != nil {
		return nil, status.Error(codes.InvalidArgument, validationErr.Error())
	}

	err = s.admin.UpdateProjectVariables(ctx, proj, req.Environment, req.Variables, req.UnsetVariables, claims.OwnerID())
	if err != nil {
		return nil, fmt.Errorf("variables updated failed with error %w", err)
	}

	vars, err := s.admin.DB.FindProjectVariables(ctx, proj.ID, nil)
	if err != nil {
		return nil, err
	}
	resp := &adminv1.UpdateProjectVariablesResponse{}
	for _, v := range vars {
		resp.Variables = append(resp.Variables, projectVariableToDTO(v))
	}
	return resp, nil
}

func (s *Server) ListProjectMemberUsers(ctx context.Context, req *adminv1.ListProjectMemberUsersRequest) (*adminv1.ListProjectMemberUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProjectMembers && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read project members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	var roleID string
	if req.Role != "" {
		role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
		if err != nil {
			return nil, err
		}
		roleID = role.ID
	}

	members, err := s.admin.DB.FindProjectMemberUsers(ctx, proj.OrganizationID, proj.ID, roleID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Email)
	}

	dtos := make([]*adminv1.ProjectMemberUser, len(members))
	for i, member := range members {
		dtos[i] = projMemberUserToPB(member)
	}

	return &adminv1.ListProjectMemberUsersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListProjectInvites(ctx context.Context, req *adminv1.ListProjectInvitesRequest) (*adminv1.ListProjectInvitesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ReadProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read project members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	// get pending user invites for this project
	userInvites, err := s.admin.DB.FindProjectInvites(ctx, proj.ID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(userInvites) >= pageSize {
		nextToken = marshalPageToken(userInvites[len(userInvites)-1].Email)
	}

	invitesDtos := make([]*adminv1.ProjectInvite, len(userInvites))
	for i, invite := range userInvites {
		invitesDtos[i] = projInviteToPB(invite)
	}

	return &adminv1.ListProjectInvitesResponse{
		Invites:       invitesDtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) AddProjectMemberUser(ctx context.Context, req *adminv1.AddProjectMemberUserRequest) (*adminv1.AddProjectMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.role", req.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add project members")
	}

	// Check outstanding invites quota
	count, err := s.admin.DB.CountInvitesForOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}
	if org.QuotaOutstandingInvites >= 0 && count >= org.QuotaOutstandingInvites {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org %q can at most have %d outstanding invitations", org.Name, org.QuotaOutstandingInvites)
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if role.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	var invitedByUserID, invitedByName string
	if claims.OwnerType() == auth.OwnerTypeUser {
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if user != nil {
			invitedByUserID = user.ID
			invitedByName = user.DisplayName
		}
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}

		// Find the guest role
		guestRole, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameGuest)
		if err != nil {
			return nil, err
		}

		// Insert an organization guest invite (will fail with a constraint error if an org-level invite already exists).
		// NOTE: Not using a transaction here for simplicity. The operation is idempotent and worst-case the user becomes a guest member with no access.
		err = s.admin.DB.InsertOrganizationInvite(ctx, &database.InsertOrganizationInviteOptions{
			Email:     req.Email,
			OrgID:     proj.OrganizationID,
			RoleID:    guestRole.ID,
			InviterID: invitedByUserID,
		})
		if err != nil && !errors.Is(err, database.ErrNotUnique) {
			return nil, err
		}

		// Find the organization invite
		orgInvite, err := s.admin.DB.FindOrganizationInvite(ctx, proj.OrganizationID, req.Email)
		if err != nil {
			if errors.Is(err, ctx.Err()) {
				return nil, err
			}
			return nil, fmt.Errorf("expected but failed to find organization invite: %w", err)
		}

		// Invite user to join the project
		err = s.admin.DB.InsertProjectInvite(ctx, &database.InsertProjectInviteOptions{
			Email:       req.Email,
			OrgInviteID: orgInvite.ID,
			ProjectID:   proj.ID,
			RoleID:      role.ID,
			InviterID:   invitedByUserID,
		})
		// continue sending an email if an invitation entry already exists
		if err != nil && !errors.Is(err, database.ErrNotUnique) {
			return nil, err
		}

		// Send invitation email
		err = s.admin.Email.SendProjectInvite(&email.ProjectInvite{
			ToEmail:       req.Email,
			ToName:        "",
			AcceptURL:     s.admin.URLs.WithCustomDomain(org.CustomDomain).ProjectInviteAccept(org.Name, proj.Name),
			OrgName:       org.Name,
			ProjectName:   proj.Name,
			RoleName:      role.Name,
			InvitedByName: invitedByName,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &adminv1.AddProjectMemberUserResponse{
			PendingSignup: true,
		}, nil
	}

	// Add the user to the project.
	err = s.admin.InsertProjectMemberUser(ctx, proj.OrganizationID, proj.ID, user.ID, role.ID, nil)
	if err != nil {
		if !errors.Is(err, database.ErrNotUnique) {
			return nil, err
		}
		// Even if the user is already a member, we continue to send the email again. Maybe they missed it the first time.
	}

	err = s.admin.Email.SendProjectAddition(&email.ProjectAddition{
		ToEmail:       req.Email,
		ToName:        "",
		OpenURL:       s.admin.URLs.WithCustomDomain(org.CustomDomain).Project(org.Name, proj.Name),
		OrgName:       org.Name,
		ProjectName:   proj.Name,
		RoleName:      role.Name,
		InvitedByName: invitedByName,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.AddProjectMemberUserResponse{
		PendingSignup: false,
	}, nil
}

func (s *Server) RemoveProjectMemberUser(ctx context.Context, req *adminv1.RemoveProjectMemberUserRequest) (*adminv1.RemoveProjectMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}

		// Only admins can remove pending invites.
		// NOTE: If we change invites to accept/decline (instead of auto-accept on signup), we need to revisit this.
		claims := auth.GetClaims(ctx)
		if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to remove project members")
		}

		// Check if there is a pending invite
		invite, err := s.admin.DB.FindProjectInvite(ctx, proj.ID, req.Email)
		if err != nil {
			return nil, err
		}

		err = s.admin.DB.DeleteProjectInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}
		return &adminv1.RemoveProjectMemberUserResponse{}, nil
	}

	// The caller must either have ManageProjectMembers permission or be the user being removed.
	claims := auth.GetClaims(ctx)
	isManager := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers
	isSelf := claims.OwnerType() == auth.OwnerTypeUser && claims.OwnerID() == user.ID
	if !isManager && !isSelf {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove project members")
	}
	if !isSelf {
		currentRole, err := s.admin.DB.FindProjectMemberUserRole(ctx, proj.ID, user.ID)
		if err != nil {
			return nil, err
		}
		if currentRole.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
			return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to remove an admin")
		}
	}

	err = s.admin.DB.DeleteProjectMemberUser(ctx, proj.ID, user.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveProjectMemberUserResponse{}, nil
}

func (s *Server) SetProjectMemberUserRole(ctx context.Context, req *adminv1.SetProjectMemberUserRoleRequest) (*adminv1.SetProjectMemberUserRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.role", req.Role),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project member roles")
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if role.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}
		// Check if there is a pending invite for this user
		invite, err := s.admin.DB.FindProjectInvite(ctx, proj.ID, req.Email)
		if err != nil {
			return nil, err
		}
		err = s.admin.DB.UpdateProjectInviteRole(ctx, invite.ID, role.ID)
		if err != nil {
			return nil, err
		}
		return &adminv1.SetProjectMemberUserRoleResponse{}, nil
	}

	currentRole, err := s.admin.DB.FindProjectMemberUserRole(ctx, proj.ID, user.ID)
	if err != nil {
		return nil, err
	}
	if currentRole.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to remove an admin")
	}

	err = s.admin.DB.UpdateProjectMemberUserRole(ctx, proj.ID, user.ID, role.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.SetProjectMemberUserRoleResponse{}, nil
}

func (s *Server) GetCloneCredentials(ctx context.Context, req *adminv1.GetCloneCredentialsRequest) (*adminv1.GetCloneCredentialsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject && !forceAccess {
		// neither a superuser nor can manage the project
		return nil, status.Error(codes.PermissionDenied, "does not have permission to get clone credentials")
	}

	if proj.ArchiveAssetID != nil {
		asset, err := s.admin.DB.FindAsset(ctx, *proj.ArchiveAssetID)
		if err != nil {
			return nil, err
		}
		downloadURL, err := s.generateSignedDownloadURL(asset)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &adminv1.GetCloneCredentialsResponse{ArchiveDownloadUrl: downloadURL}, nil
	}

	if proj.GitRemote == nil || proj.GithubInstallationID == nil {
		return nil, status.Error(codes.FailedPrecondition, "project's repository is not managed by Rill, and it does not have a GitHub integration")
	}

	repoID, err := s.githubRepoIDForProject(ctx, proj)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.admin.Github.InstallationToken(ctx, *proj.GithubInstallationID, repoID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.GetCloneCredentialsResponse{
		GitRepoUrl:           *proj.GitRemote,
		GitUsername:          "x-access-token",
		GitPassword:          token,
		GitPasswordExpiresAt: timestamppb.New(expiresAt),
		GitSubpath:           proj.Subpath,
		GitProdBranch:        proj.ProdBranch,
		GitManagedRepo:       proj.ManagedGitRepoID != nil,
	}, nil
}

func (s *Server) RequestProjectAccess(ctx context.Context, req *adminv1.RequestProjectAccessRequest) (*adminv1.RequestProjectAccessResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	projectPermissions := claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID)
	if (req.Role != database.ProjectRoleNameAdmin && projectPermissions.ReadProject) ||
		(req.Role == database.ProjectRoleNameAdmin && projectPermissions.ManageProject) {
		return nil, status.Error(codes.InvalidArgument, "already have access to project")
	}

	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.InvalidArgument, "only users can request access")
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	existing, err := s.admin.DB.FindProjectAccessRequest(ctx, proj.ID, user.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "have already requested access to project")
	}

	accessReq, err := s.admin.DB.InsertProjectAccessRequest(ctx, &database.InsertProjectAccessRequestOptions{
		UserID:    user.ID,
		ProjectID: proj.ID,
	})
	if err != nil {
		return nil, err
	}

	admins, err := s.admin.DB.FindOrganizationMembersWithManageUsersRole(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	for _, u := range admins {
		err = s.admin.Email.SendProjectAccessRequest(&email.ProjectAccessRequest{
			ToEmail:     u.Email,
			ToName:      u.DisplayName,
			Email:       user.Email,
			OrgName:     org.Name,
			ProjectName: proj.Name,
			Role:        req.Role,
			ApproveLink: s.admin.URLs.WithCustomDomain(org.CustomDomain).ApproveProjectAccess(org.Name, proj.Name, accessReq.ID, req.Role),
			DenyLink:    s.admin.URLs.WithCustomDomain(org.CustomDomain).DenyProjectAccess(org.Name, proj.Name, accessReq.ID),
		})
		if err != nil {
			return nil, err
		}
	}

	return &adminv1.RequestProjectAccessResponse{}, nil
}

func (s *Server) GetProjectAccessRequest(ctx context.Context, req *adminv1.GetProjectAccessRequestRequest) (*adminv1.GetProjectAccessRequestResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.id", req.Id),
	)

	accessReq, err := s.admin.DB.FindProjectAccessRequestByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, accessReq.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	// for now only admins can view these.
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to view project access request")
	}

	user, err := s.admin.DB.FindUser(ctx, accessReq.UserID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &adminv1.GetProjectAccessRequestResponse{Email: user.Email}, nil
}

func (s *Server) ApproveProjectAccess(ctx context.Context, req *adminv1.ApproveProjectAccessRequest) (*adminv1.ApproveProjectAccessResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.id", req.Id),
	)

	accessReq, err := s.admin.DB.FindProjectAccessRequestByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, accessReq.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project member roles")
	}

	user, err := s.admin.DB.FindUser(ctx, accessReq.UserID)
	if err != nil {
		return nil, err
	}

	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	if role.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	ok, err := s.admin.DB.CheckUserIsAProjectMember(ctx, user.ID, proj.ID)
	if err != nil {
		return nil, err
	}

	if ok {
		// User is already a project member, update the role.
		err = s.admin.DB.UpdateProjectMemberUserRole(ctx, proj.ID, user.ID, role.ID)
		if err != nil {
			return nil, err
		}
	} else {
		// Add the user as a project member.
		err = s.admin.InsertProjectMemberUser(ctx, proj.OrganizationID, proj.ID, user.ID, role.ID, nil)
		if err != nil {
			return nil, err
		}
	}

	// Remove the access request.
	err = s.admin.DB.DeleteProjectAccessRequest(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	err = s.admin.Email.SendProjectAccessGranted(&email.ProjectAccessGranted{
		ToEmail:     user.Email,
		ToName:      user.DisplayName,
		OpenURL:     s.admin.URLs.WithCustomDomain(org.CustomDomain).Project(org.Name, proj.Name),
		OrgName:     org.Name,
		ProjectName: proj.Name,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.ApproveProjectAccessResponse{}, nil
}

func (s *Server) DenyProjectAccess(ctx context.Context, req *adminv1.DenyProjectAccessRequest) (*adminv1.DenyProjectAccessResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.id", req.Id),
	)

	accessReq, err := s.admin.DB.FindProjectAccessRequestByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err := s.admin.DB.FindProject(ctx, accessReq.ProjectID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set project member roles")
	}

	user, err := s.admin.DB.FindUser(ctx, accessReq.UserID)
	if err != nil {
		return nil, err
	}
	org, err := s.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	// remove the invitation
	err = s.admin.DB.DeleteProjectAccessRequest(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	err = s.admin.Email.SendProjectAccessRejected(&email.ProjectAccessRejected{
		ToEmail:     user.Email,
		ToName:      user.DisplayName,
		OrgName:     org.Name,
		ProjectName: proj.Name,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.DenyProjectAccessResponse{}, nil
}

// getAndCheckGithubInstallationID returns a valid installation ID iff app is installed and user is a collaborator of the repo
func (s *Server) getAndCheckGithubInstallationID(ctx context.Context, gitRemote, userID string) (repoID, installationID int64, err error) {
	// Get Github installation ID for the repo
	installationID, err = s.admin.GetGithubInstallation(ctx, gitRemote)
	if err != nil {
		if errors.Is(err, admin.ErrGithubInstallationNotFound) {
			return 0, 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to %q", gitRemote)
		}

		return 0, 0, fmt.Errorf("failed to get Github installation: %w", err)
	}

	if installationID == 0 {
		return 0, 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to %q", gitRemote)
	}

	// Check that user is a collaborator on the repo
	user, err := s.admin.DB.FindUser(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	if user.GithubUsername == "" {
		return 0, 0, status.Errorf(codes.PermissionDenied, "you have not granted Rill access to your Github account")
	}

	repo, err := s.admin.LookupGithubRepoForUser(ctx, installationID, gitRemote, user.GithubUsername)
	if err != nil {
		if errors.Is(err, admin.ErrUserIsNotCollaborator) {
			return 0, 0, status.Errorf(codes.PermissionDenied, "you are not collaborator to the repo %q", gitRemote)
		}
		return 0, 0, err
	}

	return repo.GetID(), installationID, nil
}

// SudoUpdateTags updates the tags for a project in organization for superusers
func (s *Server) SudoUpdateAnnotations(ctx context.Context, req *adminv1.SudoUpdateAnnotationsRequest) (*adminv1.SudoUpdateAnnotationsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Org),
		attribute.String("args.project", req.Project),
		attribute.Int("args.annotations", len(req.Annotations)),
	)

	// Check the request is made by a superuser
	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "not authorized to update annotations")
	}

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	proj, err = s.admin.UpdateProject(ctx, proj, &database.UpdateProjectOptions{
		Name:                 proj.Name,
		Description:          proj.Description,
		Public:               proj.Public,
		DirectoryName:        proj.DirectoryName,
		ArchiveAssetID:       proj.ArchiveAssetID,
		GitRemote:            proj.GitRemote,
		GithubInstallationID: proj.GithubInstallationID,
		GithubRepoID:         proj.GithubRepoID,
		ManagedGitRepoID:     proj.ManagedGitRepoID,
		ProdVersion:          proj.ProdVersion,
		ProdBranch:           proj.ProdBranch,
		Subpath:              proj.Subpath,
		ProdDeploymentID:     proj.ProdDeploymentID,
		ProdSlots:            proj.ProdSlots,
		ProdTTLSeconds:       proj.ProdTTLSeconds,
		DevSlots:             proj.DevSlots,
		DevTTLSeconds:        proj.DevTTLSeconds,
		Provisioner:          proj.Provisioner,
		Annotations:          req.Annotations,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateAnnotationsResponse{
		Project: s.projToDTO(proj, req.Org),
	}, nil
}

func (s *Server) CreateProjectWhitelistedDomain(ctx context.Context, req *adminv1.CreateProjectWhitelistedDomainRequest) (*adminv1.CreateProjectWhitelistedDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.domain", req.Domain),
		attribute.String("args.role", req.Role),
	)

	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	if !claims.Superuser(ctx) {
		if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject {
			return nil, status.Error(codes.PermissionDenied, "only proj admins can add whitelisted domain")
		}
		// check if the user's domain matches the whitelist domain
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(user.Email, "@"+req.Domain) {
			return nil, status.Error(codes.PermissionDenied, "Domain name doesn’t match verified email domain. Please contact Rill support.")
		}

		if publicemail.IsPublic(req.Domain) {
			return nil, status.Errorf(codes.InvalidArgument, "Public Domain %s cannot be whitelisted", req.Domain)
		}
	}

	role, err := s.admin.DB.FindProjectRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}
	if role.Admin && !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProjectAdmins {
		return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to assign an admin role")
	}

	// find existing users belonging to the whitelisted domain to the project
	users, err := s.admin.DB.FindUsersByEmailPattern(ctx, "%@"+req.Domain, "", math.MaxInt)
	if err != nil {
		return nil, err
	}

	// filter out users who are already members of the project
	newUsers := make([]*database.User, 0)
	for _, user := range users {
		// check if user is already a member of the project
		exists, err := s.admin.DB.CheckUserIsAProjectMember(ctx, user.ID, proj.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			newUsers = append(newUsers, user)
		}
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = s.admin.DB.InsertProjectWhitelistedDomain(ctx, &database.InsertProjectWhitelistedDomainOptions{
		ProjectID:     proj.ID,
		ProjectRoleID: role.ID,
		Domain:        req.Domain,
	})
	if err != nil {
		return nil, err
	}

	for _, user := range newUsers {
		// Add the user to the project.
		err = s.admin.InsertProjectMemberUser(ctx, proj.OrganizationID, proj.ID, user.ID, role.ID, nil)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateProjectWhitelistedDomainResponse{}, nil
}

func (s *Server) RemoveProjectWhitelistedDomain(ctx context.Context, req *adminv1.RemoveProjectWhitelistedDomainRequest) (*adminv1.RemoveProjectWhitelistedDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.domain", req.Domain),
	)

	claims := auth.GetClaims(ctx)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	if !(claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only project admins can remove whitelisted domain")
	}

	invite, err := s.admin.DB.FindProjectWhitelistedDomain(ctx, proj.ID, req.Domain)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteProjectWhitelistedDomain(ctx, invite.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveProjectWhitelistedDomainResponse{}, nil
}

func (s *Server) ListProjectWhitelistedDomains(ctx context.Context, req *adminv1.ListProjectWhitelistedDomainsRequest) (*adminv1.ListProjectWhitelistedDomainsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	if !(claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only project admins can list whitelisted domains")
	}

	domains, err := s.admin.DB.FindProjectWhitelistedDomainForProjectWithJoinedRoleNames(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*adminv1.WhitelistedDomain, len(domains))
	for i, domain := range domains {
		dtos[i] = &adminv1.WhitelistedDomain{
			Domain: domain.Domain,
			Role:   domain.RoleName,
		}
	}

	return &adminv1.ListProjectWhitelistedDomainsResponse{Domains: dtos}, nil
}

func (s *Server) RedeployProject(ctx context.Context, req *adminv1.RedeployProjectRequest) (*adminv1.RedeployProjectResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProd && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check if org has blocking billing errors
	err = s.admin.CheckBlockingBillingErrors(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var depl *database.Deployment
	if proj.ProdDeploymentID != nil {
		depl, err = s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	_, err = s.admin.RedeployProject(ctx, proj, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.RedeployProjectResponse{}, nil
}

func (s *Server) HibernateProject(ctx context.Context, req *adminv1.HibernateProjectRequest) (*adminv1.HibernateProjectResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
	)

	proj, err := s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProject && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "not allowed to manage project")
	}

	_, err = s.admin.HibernateProject(ctx, proj)
	if err != nil {
		return nil, fmt.Errorf("failed to hibernate project: %w", err)
	}

	return &adminv1.HibernateProjectResponse{}, nil
}

// Deprecated: See api.proto for details.
func (s *Server) TriggerRedeploy(ctx context.Context, req *adminv1.TriggerRedeployRequest) (*adminv1.TriggerRedeployResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.organization", req.Org),
		attribute.String("args.project", req.Project),
		attribute.String("args.deployment_id", req.DeploymentId),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Org)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check if org has blocking billing errors
	err = s.admin.CheckBlockingBillingErrors(ctx, org.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// For backwards compatibility, this RPC supports passing either DeploymentId or Organization+Project names
	var proj *database.Project
	var depl *database.Deployment
	if req.DeploymentId != "" {
		var err error
		depl, err = s.admin.DB.FindDeployment(ctx, req.DeploymentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		proj, err = s.admin.DB.FindProject(ctx, depl.ProjectID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	} else {
		var err error
		proj, err = s.admin.DB.FindProjectByName(ctx, req.Org, req.Project)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if proj.ProdDeploymentID != nil {
			depl, err = s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	claims := auth.GetClaims(ctx)
	if !claims.ProjectPermissions(ctx, proj.OrganizationID, proj.ID).ManageProd {
		return nil, status.Error(codes.PermissionDenied, "does not have permission to manage deployment")
	}

	_, err = s.admin.RedeployProject(ctx, proj, depl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.TriggerRedeployResponse{}, nil
}

func (s *Server) projToDTO(p *database.Project, orgName string) *adminv1.Project {
	return &adminv1.Project{
		Id:               p.ID,
		Name:             p.Name,
		OrgId:            p.OrganizationID,
		OrgName:          orgName,
		Description:      p.Description,
		Public:           p.Public,
		CreatedByUserId:  safeStr(p.CreatedByUserID),
		DirectoryName:    p.DirectoryName,
		Provisioner:      p.Provisioner,
		ProdVersion:      p.ProdVersion,
		ProdSlots:        int64(p.ProdSlots),
		ProdBranch:       p.ProdBranch,
		Subpath:          p.Subpath,
		GitRemote:        safeStr(p.GitRemote),
		ManagedGitId:     safeStr(p.ManagedGitRepoID),
		ArchiveAssetId:   safeStr(p.ArchiveAssetID),
		ProdDeploymentId: safeStr(p.ProdDeploymentID),
		ProdTtlSeconds:   safeInt64(p.ProdTTLSeconds),
		FrontendUrl:      s.admin.URLs.Project(orgName, p.Name),
		Annotations:      p.Annotations,
		CreatedOn:        timestamppb.New(p.CreatedOn),
		UpdatedOn:        timestamppb.New(p.UpdatedOn),
	}
}

func (s *Server) hasAssetUsagePermission(ctx context.Context, id, orgID, ownerID string) bool {
	asset, err := s.admin.DB.FindAsset(ctx, id)
	if err != nil {
		return false
	}
	return asset.OrganizationID != nil && *asset.OrganizationID == orgID && asset.OwnerID == ownerID
}

func (s *Server) githubOptsForRemote(ctx context.Context, orgID, branch string, userID *string, gitRemote string) (githubRepoID, instID *int64, mgdGitRepoID *string, prodBranch string, resErr error) {
	isMgdGitRepo := true
	mgdGitRepo, err := s.admin.DB.FindManagedGitRepo(ctx, gitRemote)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, nil, nil, "", err
		}
		isMgdGitRepo = false
	}
	if isMgdGitRepo {
		// rill managed git repo
		if mgdGitRepo.OrgID == nil || orgID != *mgdGitRepo.OrgID {
			return nil, nil, nil, "", status.Error(codes.PermissionDenied, "not allowed to access this managed git repo")
		}
		id, err := s.admin.Github.ManagedOrgInstallationID()
		if err != nil {
			return nil, nil, nil, "", err
		}

		// fetch github repo id from github
		// ideally this can be stored in managed git repo table but it is fine to fetch it from github during project creation/updation
		c := s.admin.Github.InstallationClient(id, nil)
		account, repo, ok := gitutil.SplitGithubRemote(gitRemote)
		if !ok {
			return nil, nil, nil, "", status.Error(codes.InvalidArgument, "invalid github url")
		}
		ghRepo, _, err := c.Repositories.Get(ctx, account, repo)
		if err != nil {
			return nil, nil, nil, "", status.Error(codes.InvalidArgument, "failed to get github repo")
		}

		if branch == "" {
			branch = ghRepo.GetDefaultBranch()
		}
		return ghRepo.ID, &id, &mgdGitRepo.ID, branch, nil
	}
	// User managed github projects must be configured by a user so we can ensure that they're allowed to access the repo.
	if userID == nil {
		return nil, nil, nil, "", status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Check Github app is installed and caller has access on the repo
	ghRepoID, installationID, err := s.getAndCheckGithubInstallationID(ctx, gitRemote, *userID)
	if err != nil {
		return nil, nil, nil, "", err
	}
	return &ghRepoID, &installationID, nil, branch, nil
}

// githubRepoIDForProject returns the github repo id for a project stored in the database.
// For older projects this may be nil since the github repo id was not stored.
// This function fetches the ID from github API and stores it in the database.
// It assumes that the project is connected to github and necessary checks have been done by the caller.
func (s *Server) githubRepoIDForProject(ctx context.Context, p *database.Project) (int64, error) {
	if p.GithubRepoID != nil {
		return *p.GithubRepoID, nil
	}

	if p.GithubInstallationID == nil {
		return 0, fmt.Errorf("project %q is not connected to github", p.Name)
	}

	client := s.admin.Github.InstallationClient(*p.GithubInstallationID, nil)
	account, repo, ok := gitutil.SplitGithubRemote(*p.GitRemote)
	if !ok {
		return 0, status.Error(codes.InvalidArgument, "invalid github url")
	}

	ghRepo, _, err := client.Repositories.Get(ctx, account, repo)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "failed to get github repo")
	}
	id := ghRepo.GetID()
	_, err = s.admin.DB.UpdateProject(ctx, p.ID, &database.UpdateProjectOptions{
		Name:                 p.Name,
		Description:          p.Description,
		Public:               p.Public,
		DirectoryName:        p.DirectoryName,
		ArchiveAssetID:       p.ArchiveAssetID,
		GitRemote:            p.GitRemote,
		GithubInstallationID: p.GithubInstallationID,
		GithubRepoID:         &id,
		ManagedGitRepoID:     p.ManagedGitRepoID,
		ProdVersion:          p.ProdVersion,
		ProdBranch:           p.ProdBranch,
		Subpath:              p.Subpath,
		ProdDeploymentID:     p.ProdDeploymentID,
		ProdSlots:            p.ProdSlots,
		ProdTTLSeconds:       p.ProdTTLSeconds,
		DevSlots:             p.DevSlots,
		DevTTLSeconds:        p.DevTTLSeconds,
		Provisioner:          p.Provisioner,
		Annotations:          p.Annotations,
	})
	if err != nil {
		return 0, status.Error(codes.Internal, "failed to update project with github repo id")
	}
	return id, nil
}

func deploymentToDTO(d *database.Deployment) *adminv1.Deployment {
	var s adminv1.DeploymentStatus
	switch d.Status {
	case database.DeploymentStatusUnspecified:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_UNSPECIFIED
	case database.DeploymentStatusPending:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_PENDING
	case database.DeploymentStatusUpdating:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_UPDATING
	case database.DeploymentStatusRunning:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING
	case database.DeploymentStatusErrored:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_ERRORED
	case database.DeploymentStatusStopping:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPING
	case database.DeploymentStatusStopped:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPED
	case database.DeploymentStatusDeleting:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_DELETING
	case database.DeploymentStatusDeleted:
		s = adminv1.DeploymentStatus_DEPLOYMENT_STATUS_DELETED
	default:
		panic(fmt.Errorf("unhandled deployment status %d", d.Status))
	}

	return &adminv1.Deployment{
		Id:                d.ID,
		ProjectId:         d.ProjectID,
		OwnerUserId:       safeStr(d.OwnerUserID),
		Environment:       d.Environment,
		Branch:            d.Branch,
		RuntimeHost:       d.RuntimeHost,
		RuntimeInstanceId: d.RuntimeInstanceID,
		Status:            s,
		StatusMessage:     d.StatusMessage,
		CreatedOn:         timestamppb.New(d.CreatedOn),
		UpdatedOn:         timestamppb.New(d.UpdatedOn),
	}
}

func projectVariableToDTO(v *database.ProjectVariable) *adminv1.ProjectVariable {
	return &adminv1.ProjectVariable{
		Id:              v.ID,
		Name:            v.Name,
		Value:           v.Value,
		Environment:     v.Environment,
		UpdatedByUserId: safeStr(v.UpdatedByUserID),
		CreatedOn:       timestamppb.New(v.CreatedOn),
		UpdatedOn:       timestamppb.New(v.UpdatedOn),
	}
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeInt64(s *int64) int64 {
	if s == nil {
		return 0
	}
	return *s
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
