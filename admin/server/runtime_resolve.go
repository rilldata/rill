package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime/pkg/httputil"
)

// errNoProdAccess is returned when the caller does not have access to a project's production deployment.
// It is a sentinel because callers own the response and decide whether to emit an OAuth challenge.
var errNoProdAccess = errors.New("does not have permission to access the production deployment")

// resolveDeploymentForOrgAndProject resolves an org and project name to the deployment to serve requests from.
// If branch is empty, it resolves the project's primary deployment.
func (s *Server) resolveDeploymentForOrgAndProject(ctx context.Context, org, project, branch string) (*database.Project, *database.Deployment, error) {
	proj, err := s.admin.DB.FindProjectByName(ctx, org, project)
	if err != nil {
		return nil, nil, httputil.Error(http.StatusBadRequest, err)
	}

	if branch == "" {
		if proj.PrimaryDeploymentID == nil {
			return nil, nil, httputil.Errorf(http.StatusBadRequest, "no prod deployment for project")
		}
		depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
		if err != nil {
			return nil, nil, httputil.Error(http.StatusBadRequest, err)
		}
		return proj, depl, nil
	}

	depls, err := s.admin.DB.FindDeploymentsForProject(ctx, proj.ID, "", branch)
	if err != nil {
		return nil, nil, httputil.Error(http.StatusBadRequest, err)
	}
	if len(depls) == 0 {
		return nil, nil, httputil.Errorf(http.StatusBadRequest, "no deployment for branch %q", branch)
	}
	return proj, depls[0], nil // At most one deployment per branch is allowed
}

// issueEphemeralRuntimeToken checks that the caller can read a project's production deployment,
// then issues a runtime JWT for it similar to the one that could be obtained by calling GetProject.
// It returns errNoProdAccess if the caller does not have access.
func (s *Server) issueEphemeralRuntimeToken(ctx context.Context, proj *database.Project, depl *database.Deployment, ttl time.Duration) (string, error) {
	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, depl.ProjectID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}
	if !permissions.ReadProd {
		return "", errNoProdAccess
	}

	return s.issueRuntimeToken(ctx, &issueRuntimeTokenOptions{
		project:            proj,
		deployment:         depl,
		projectPermissions: permissions,
		forOwner:           true,
		ttl:                ttl,
	})
}

// runtimeHTTPHost returns the host to send HTTP requests to for a deployment.
// NOTE: In production, the runtime host serves both the HTTP and gRPC servers.
// But in development, the two are presently on different ports, and depl.RuntimeHost is that of the gRPC server.
// Until we get both servers on the same port in development, this hack rewrites the runtime host to the HTTP server.
func runtimeHTTPHost(runtimeHost string) string {
	if !strings.HasPrefix(runtimeHost, "http://localhost:") {
		return runtimeHost
	}
	if host := os.Getenv("RILL_RUNTIME_AUTH_AUDIENCE_URL"); host != "" {
		return host
	}
	return "http://localhost:8081"
}
