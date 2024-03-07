package server

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
)

// runtimeProxyForOrgAndProject proxies a request to the runtime service for a specific project.
// This provides a way to directly query a project's runtime on a stable URL without needing to call GetProject or GetDeploymentCredentials to discover the runtime URL.
// If the request is made using an Authorization header or cookie recognized by the admin service,
// the proxied request is made with a newly minted JWT similar to the one that could be obtained by calling GetProject.
// If the Authorization header of the request is not recognized by the admin service, it is proxied through to the runtime service.
func (s *Server) runtimeProxyForOrgAndProject(w http.ResponseWriter, r *http.Request) error {
	// Get args from URL path components
	org := r.PathValue("org")
	project := r.PathValue("project")
	proxyPath := r.PathValue("path")

	// Find the production deployment for the project we're proxying to
	proj, err := s.admin.DB.FindProjectByName(r.Context(), org, project)
	if err != nil {
		return httputil.Error(http.StatusBadRequest, err)
	}
	if proj.ProdDeploymentID == nil {
		return httputil.Errorf(http.StatusBadRequest, "no prod deployment for project")
	}
	depl, err := s.admin.DB.FindDeployment(r.Context(), *proj.ProdDeploymentID)
	if err != nil {
		return httputil.Error(http.StatusBadRequest, err)
	}

	// Get or issue a JWT to use for the proxied request.
	var jwt string
	claims := auth.GetClaims(r.Context())
	switch claims.OwnerType() {
	case auth.OwnerTypeAnon:
		// If the client is not authenticated with the admin service, we just proxy the contents of the Authorization header to the runtime (if any).
		// Note that the authorization middleware for this handler is set to be "lenient",
		// which means it will still invoke this handler even if the Authorization header contains a token that is not valid for the admin service.
		authorizationHeader := r.Header.Get("Authorization")
		if len(authorizationHeader) >= 6 && strings.EqualFold(authorizationHeader[0:6], "bearer") {
			jwt = strings.TrimSpace(authorizationHeader[6:])
		}
	case auth.OwnerTypeUser, auth.OwnerTypeService:
		// If the client is authenticated with the admin service, we issue a new ephemeral runtime JWT.
		// The JWT should have the same permissions/configuration as one they would get by calling AdminService.GetProject.

		permissions := claims.ProjectPermissions(r.Context(), proj.OrganizationID, depl.ProjectID)
		if !permissions.ReadProd {
			return httputil.Errorf(http.StatusForbidden, "does not have permission to access the production deployment")
		}

		var attr map[string]any
		if claims.OwnerType() == auth.OwnerTypeUser {
			attr, err = s.jwtAttributesForUser(r.Context(), claims.OwnerID(), proj.OrganizationID, permissions)
			if err != nil {
				return httputil.Error(http.StatusInternalServerError, err)
			}
		}

		jwt, err = s.issuer.NewToken(runtimeauth.TokenOptions{
			AudienceURL: depl.RuntimeAudience,
			Subject:     claims.OwnerID(),
			TTL:         runtimeAccessTokenDefaultTTL,
			InstancePermissions: map[string][]runtimeauth.Permission{
				depl.RuntimeInstanceID: {
					// TODO: Remove ReadProfiling and ReadRepo (may require frontend changes)
					runtimeauth.ReadObjects,
					runtimeauth.ReadMetrics,
					runtimeauth.ReadProfiling,
					runtimeauth.ReadRepo,
					runtimeauth.ReadAPI,
				},
			},
			Attributes: attr,
		})
		if err != nil {
			return httputil.Error(http.StatusInternalServerError, err)
		}
	default:
		return httputil.Errorf(http.StatusBadRequest, "runtime proxy not available for owner type %q", claims.OwnerType())
	}

	// Track usage of the deployment
	s.admin.Used.Deployment(depl.ID)

	// Determine runtime host.
	// NOTE: In production, the runtime host serves both the HTTP and gRPC servers.
	// But in development, the two are presently on different ports, and depl.RuntimeHost is that of the gRPC server.
	// Until we get both servers on the same port in development, this hack rewrites the runtime host to the HTTP server.
	runtimeHost := depl.RuntimeHost
	if strings.HasPrefix(runtimeHost, "http://localhost:") {
		runtimeHost = os.Getenv("RILL_RUNTIME_AUTH_AUDIENCE_URL")
		if runtimeHost == "" {
			runtimeHost = "http://localhost:8081"
		}
	}

	// Create the URL to proxy to by prepending `/v1/instances/{instanceID}` to the proxy path.
	proxyURL, err := url.JoinPath(runtimeHost, "/v1/instances", depl.RuntimeInstanceID, proxyPath)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// Create the proxied request.
	req, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	for k, v := range r.Header {
		req.Header.Add(k, v[0])
	}

	// Override the authorization header with the JWT (note use of Set instead of Add).
	if jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	} else {
		req.Header.Del("Authorization")
	}

	// Send the proxied request using http.DefaultClient. The default client automatically handles caching/pooling of TCP connections.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	defer res.Body.Close()

	// Copy the proxied response to the original response writer
	outHeader := w.Header()
	for k, v := range res.Header {
		for _, vv := range v {
			outHeader.Add(k, vv)
		}
	}
	w.WriteHeader(res.StatusCode)
	_, err = io.Copy(w, res.Body)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}
