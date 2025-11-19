package server

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
)

// runtimeProxyAccessTokenTTL is the TTL for tokens minted by the runtime proxy.
// Since streaming connections and MCP SSE connections can be long-lived, we set this to a long duration.
const runtimeProxyAccessTokenTTL = 24 * time.Hour

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
	proxyRawQuery := r.URL.RawQuery

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

	// Prepare a JWT to use for the proxied request.
	// We support three scenarios:
	// 1. Passing a runtime JWT directly.
	// 2. Using admin service authentication, which requires us to issue a new ephemeral runtime JWT for the proxied request.
	// 3. Accessing public projects anonymously, which also requires us to issue a new ephemeral runtime JWT for the proxied request.
	var jwt string

	// We support passing a runtime JWT directly.
	// Since we use HTTPMiddlewareLenient for this handler, it's invoked even if the Authorization header contains a token that is not valid for the admin service.
	claims := auth.GetClaims(r.Context())
	if claims.OwnerType() == auth.OwnerTypeAnon {
		authorizationHeader := r.Header.Get("Authorization")
		if len(authorizationHeader) >= 6 && strings.EqualFold(authorizationHeader[0:6], "bearer") {
			jwt = strings.TrimSpace(authorizationHeader[6:])
		}
	}
	// If a direct JWT was not provided, we rely on admin service auth to issue a new ephemeral runtime JWT for the proxied request.
	// TODO: This mirrors logic in GetProject. Consider refactoring to avoid duplication.
	if jwt == "" {
		permissions := claims.ProjectPermissions(r.Context(), proj.OrganizationID, depl.ProjectID)
		if proj.Public {
			permissions.ReadProject = true
			permissions.ReadProd = true
		}
		if !permissions.ReadProd {
			if claims.OwnerType() == auth.OwnerTypeAnon {
				// This means no token was provided, so return instructions for how to initiate an OAuth flow.
				// This is currently used by MCP clients that authenticate with OAuth.
				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=%q", s.admin.URLs.OAuthProtectedResourceMetadata()))
			}
			return httputil.Errorf(http.StatusUnauthorized, "does not have permission to access the production deployment")
		}

		var attr map[string]any
		switch claims.OwnerType() {
		case auth.OwnerTypeAnon:
			// No attributes
		case auth.OwnerTypeUser:
			attr, err = s.jwtAttributesForUser(r.Context(), claims.OwnerID(), proj.OrganizationID, permissions)
			if err != nil {
				return httputil.Error(http.StatusInternalServerError, err)
			}
		case auth.OwnerTypeService:
			attr, err = s.jwtAttributesForService(r.Context(), claims.OwnerID(), permissions)
			if err != nil {
				return httputil.Error(http.StatusInternalServerError, err)
			}
		default:
			return httputil.Errorf(http.StatusBadRequest, "runtime proxy not available for owner type %q", claims.OwnerType())
		}

		instancePermissions := []runtime.Permission{
			runtime.ReadObjects,
			runtime.ReadMetrics,
			runtime.ReadAPI,
			runtime.UseAI,
		}
		if permissions.ManageProject {
			instancePermissions = append(instancePermissions, runtime.EditTrigger)
		}

		jwt, err = s.issuer.NewToken(runtimeauth.TokenOptions{
			AudienceURL: depl.RuntimeAudience,
			Subject:     claims.OwnerID(),
			TTL:         runtimeProxyAccessTokenTTL,
			InstancePermissions: map[string][]runtime.Permission{
				depl.RuntimeInstanceID: instancePermissions,
			},
			Attributes: attr,
		})
		if err != nil {
			return httputil.Error(http.StatusInternalServerError, err)
		}
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
	proxyURL, err := url.Parse(runtimeHost)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	proxyURL = proxyURL.JoinPath("/v1/instances", depl.RuntimeInstanceID, proxyPath)
	proxyURL.RawQuery = proxyRawQuery

	// Create the proxied request.
	req, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL.String(), r.Body)
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

	// Add the X-Original-URI header to preserve the original request URI.
	// This enables the runtime to know the runtime proxy path that was used.
	req.Header.Set("X-Original-URI", r.RequestURI)

	// Send the proxied request using http.DefaultClient. The default client automatically handles caching/pooling of TCP connections.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	defer res.Body.Close()

	// Copy response headers except from "Access-Control-Allow-Origin" (which is also added by the admin server), thus causing browser CORS errors.
	outHeader := w.Header()
	for k, v := range res.Header {
		if strings.EqualFold(k, "Access-Control-Allow-Origin") {
			continue
		}
		for _, vv := range v {
			outHeader.Add(k, vv)
		}
	}
	w.WriteHeader(res.StatusCode)

	// For SSE responses, we need to flush eagerly
	if res.Header.Get("Content-Type") == "text/event-stream" {
		flusher, ok := w.(http.Flusher)
		if !ok {
			return httputil.Error(http.StatusInternalServerError, fmt.Errorf("streaming not supported"))
		}

		// Use a larger buffer for better performance
		reader := bufio.NewReaderSize(res.Body, 4096)
		buffer := make([]byte, 4096)

		for {
			n, err := reader.Read(buffer)
			if n > 0 {
				if _, writeErr := w.Write(buffer[:n]); writeErr != nil {
					return writeErr
				}
				flusher.Flush()
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	} else {
		_, err = io.Copy(w, res.Body)
		if err != nil {
			return httputil.Error(http.StatusInternalServerError, err)
		}
	}

	return nil
}
