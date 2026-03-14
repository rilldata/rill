package server

import (
	"encoding/json"
	"net/http"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	chdriver "github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rs/cors"
)

type chcLookupRequest struct {
	KeyID     string `json:"key_id"`
	KeySecret string `json:"key_secret"`
	Host      string `json:"host"`
	// Org and Project are optional; when set, the lookup persists cluster info on the project.
	Org     string `json:"org"`
	Project string `json:"project"`
}

type chcLookupResponse struct {
	ServiceName   string  `json:"service_name"`
	Status        string  `json:"status"`
	CloudProvider string  `json:"cloud_provider"`
	Region        string  `json:"region"`
	Tier          string  `json:"tier"`
	IdleScaling   bool    `json:"idle_scaling"`
	MinMemoryGB   float64 `json:"min_memory_gb"`
	MaxMemoryGB   float64 `json:"max_memory_gb"`
	NumReplicas   int     `json:"num_replicas"`
	RillMinSlots  int     `json:"rill_min_slots"`
}

func (s *Server) registerClickHouseCloudEndpoints(mux *http.ServeMux) {
	corsMiddleware := cors.New(newCORSOptions(s.opts.AllowedOrigins, true)).Handler
	handler := observability.Middleware("admin", s.logger, corsMiddleware(s.authenticator.HTTPMiddleware(httputil.Handler(s.clickhouseCloudLookup))))
	observability.MuxHandle(mux, "/v1/clickhouse-cloud/lookup", handler)
}

func (s *Server) clickhouseCloudLookup(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return nil
	}

	var req chcLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil
	}

	if req.KeyID == "" || req.KeySecret == "" || req.Host == "" {
		http.Error(w, "key_id, key_secret, and host are required", http.StatusBadRequest)
		return nil
	}

	client := chdriver.NewCloudAPIClient(req.KeyID, req.KeySecret)
	if client == nil {
		http.Error(w, "failed to create CHC API client", http.StatusInternalServerError)
		return nil
	}

	info, err := client.FindServiceByHost(r.Context(), req.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return nil
	}

	// info.MaxMemoryGB is already per-replica (from minReplicaMemoryGb/maxReplicaMemoryGb)
	minSlots := admin.CHCMinSlotsForMemory(info.MaxMemoryGB)

	// Persist cluster info on the project if org/project are provided
	if req.Org != "" && req.Project != "" {
		proj, err := s.admin.DB.FindProjectByName(r.Context(), req.Org, req.Project)
		if err == nil {
			clusterSize := info.MaxMemoryGB
			minSlotsInt64 := int64(minSlots)
			_, _ = s.admin.DB.UpdateProject(r.Context(), proj.ID, &database.UpdateProjectOptions{
				Name:                 proj.Name,
				Description:          proj.Description,
				Public:               proj.Public,
				DirectoryName:        proj.DirectoryName,
				ArchiveAssetID:       proj.ArchiveAssetID,
				GitRemote:            proj.GitRemote,
				GithubInstallationID: proj.GithubInstallationID,
				GithubRepoID:         proj.GithubRepoID,
				ManagedGitRepoID:     proj.ManagedGitRepoID,
				Subpath:              proj.Subpath,
				ProdVersion:          proj.ProdVersion,
				PrimaryBranch:        proj.PrimaryBranch,
				PrimaryDeploymentID:  proj.PrimaryDeploymentID,
				ProdSlots:            proj.ProdSlots,
				ProdTTLSeconds:       proj.ProdTTLSeconds,
				DevSlots:             proj.DevSlots,
				DevTTLSeconds:        proj.DevTTLSeconds,
				Provisioner:          proj.Provisioner,
				Annotations:          proj.Annotations,
				ChcClusterSize:       &clusterSize,
				RillMinSlots:         &minSlotsInt64,
			})
		}
	}

	resp := chcLookupResponse{
		ServiceName:   info.Name,
		Status:        info.Status,
		CloudProvider: info.CloudProvider,
		Region:        info.Region,
		Tier:          info.Tier,
		IdleScaling:   info.IdleScaling,
		MinMemoryGB:   info.MinMemoryGB,
		MaxMemoryGB:   info.MaxMemoryGB,
		NumReplicas:   info.NumReplicas,
		RillMinSlots:  minSlots,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}
