package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rilldata/rill/admin/server/auth"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) runtimeProxyForOrgAndProject(w http.ResponseWriter, r *http.Request) {
	org := r.PathValue("org")
	project := r.PathValue("project")
	proxyPath := r.PathValue("path")

	proj, err := s.admin.DB.FindProjectByName(r.Context(), org, project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if proj.ProdDeploymentID == nil {
		http.Error(w, "no prod deployment for project", http.StatusBadRequest)
		return
	}

	depl, err := s.admin.DB.FindDeployment(r.Context(), *proj.ProdDeploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims := auth.GetClaims(r.Context())
	permissions := claims.ProjectPermissions(r.Context(), proj.OrganizationID, depl.ProjectID)
	if proj.Public {
		permissions.ReadProject = true
		permissions.ReadProd = true
	}

	if !permissions.ReadProd {
		http.Error(w, "does not have permission to access the production deployment", http.StatusForbidden)
		return
	}

	s.admin.Used.Deployment(depl.ID)

	// Get the JWT (if any) to use for the proxied request.
	var jwt string
	switch claims.OwnerType() {
	// If the client is not authenticated with the admin service, we just proxy the contents of the Authorization header to the runtime (if any).
	// Note that the authorization middleware for this handler is set to be "lenient",
	// which means it will still invoke this handler even if the Authorization header contains a token that is not valid for the admin service.
	case auth.OwnerTypeAnon:
		authorizationHeader := r.Header.Get("Authorization")
		if len(authorizationHeader) >= 6 && strings.EqualFold(authorizationHeader[0:6], "bearer") {
			jwt = strings.TrimSpace(authorizationHeader[6:])
		}
	// If the client is authenticated with the admin service, we issue a new ephemeral runtime JWT.
	// The JWT should have the same permissions/configuration as one they would get by calling AdminService.GetProject.
	case auth.OwnerTypeUser, auth.OwnerTypeService:
		var attr map[string]any
		if claims.OwnerType() == auth.OwnerTypeUser {
			attr, err = s.jwtAttributesForUser(r.Context(), claims.OwnerID(), proj.OrganizationID, permissions)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
				},
			},
			Attributes: attr,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("runtime proxy not available for owner type %q", claims.OwnerType()), http.StatusBadRequest)
		return
	}

	// Create the URL to proxy to
	proxyURL, err := url.JoinPath(depl.RuntimeHost, "/v1/instances", depl.RuntimeInstanceID, proxyPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the proxied request.
	req, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for k, v := range r.Header {
		req.Header.Add(k, v[0])
	}

	// Override the authorization header with the JWT.
	req.Header.Set("Authorization", "Bearer "+jwt)

	// Send the proxied request using http.DefaultClient. The default client automatically handles caching/pooling of TCP connections.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	io.Copy(w, res.Body)
}
