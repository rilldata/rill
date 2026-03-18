package server

import (
	"encoding/json"
	"maps"
	"net/http"
	"strings"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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

	syncHandler := observability.Middleware("admin", s.logger, corsMiddleware(s.authenticator.HTTPMiddleware(httputil.Handler(s.clickhouseCloudSync))))
	observability.MuxHandle(mux, "/v1/clickhouse-cloud/sync", syncHandler)
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
				InfraSlots:           proj.InfraSlots,
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

// clickhouseCloudSync triggers a CHC status check and auto-scale for a single project.
// Allows users to immediately sync after starting/stopping their CHC cluster
// instead of waiting for the hourly billing reporter.
func (s *Server) clickhouseCloudSync(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return nil
	}

	var req struct {
		Org     string `json:"org"`
		Project string `json:"project"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil
	}
	if req.Org == "" || req.Project == "" {
		http.Error(w, "org and project are required", http.StatusBadRequest)
		return nil
	}

	proj, err := s.admin.DB.FindProjectByName(r.Context(), req.Org, req.Project)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return nil
	}
	if proj.PrimaryDeploymentID == nil {
		http.Error(w, "project has no deployment", http.StatusBadRequest)
		return nil
	}

	depl, err := s.admin.DB.FindDeployment(r.Context(), *proj.PrimaryDeploymentID)
	if err != nil || depl.Status != database.DeploymentStatusRunning {
		http.Error(w, "deployment not available", http.StatusBadRequest)
		return nil
	}

	// Get the connector host from the runtime instance
	rt, err := s.admin.OpenRuntimeClient(depl)
	if err != nil {
		http.Error(w, "failed to connect to runtime", http.StatusInternalServerError)
		return nil
	}
	defer rt.Close()

	instance, err := rt.GetInstance(r.Context(), &runtimev1.GetInstanceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Sensitive:  true,
	})
	if err != nil {
		http.Error(w, "failed to get instance", http.StatusInternalServerError)
		return nil
	}

	var connHost string
	for _, conn := range instance.Instance.ProjectConnectors {
		if conn.Name != instance.Instance.OlapConnector || conn.Config == nil {
			continue
		}
		if v, ok := conn.Config.Fields["resolved_host"]; ok {
			connHost = v.GetStringValue()
		}
		if connHost == "" {
			if v, ok := conn.Config.Fields["host"]; ok {
				connHost = v.GetStringValue()
			}
		}
		break
	}
	if connHost == "" || !strings.Contains(strings.ToLower(connHost), ".clickhouse.cloud") {
		http.Error(w, "not a ClickHouse Cloud project", http.StatusBadRequest)
		return nil
	}

	// Fetch API keys from project variables
	env := "prod"
	vars, err := s.admin.DB.FindProjectVariables(r.Context(), proj.ID, &env)
	if err != nil {
		http.Error(w, "failed to read project variables", http.StatusInternalServerError)
		return nil
	}
	var keyID, keySecret string
	for _, v := range vars {
		switch v.Name {
		case "CLICKHOUSE_CLOUD_API_KEY_ID":
			keyID = v.Value
		case "CLICKHOUSE_CLOUD_API_KEY_SECRET":
			keySecret = v.Value
		}
	}

	client := chdriver.NewCloudAPIClient(keyID, keySecret)
	if client == nil {
		http.Error(w, "no CHC API keys configured", http.StatusBadRequest)
		return nil
	}

	info, err := client.FindServiceByHost(r.Context(), connHost)
	if err != nil {
		http.Error(w, "CHC API call failed: "+err.Error(), http.StatusBadGateway)
		return nil
	}

	// Update cluster size/min slots
	minSlots := admin.CHCMinSlotsForMemory(info.MaxMemoryGB)
	minSlotsInt64 := int64(minSlots)
	if proj.ChcClusterSize == nil || *proj.ChcClusterSize != info.MaxMemoryGB ||
		proj.RillMinSlots == nil || *proj.RillMinSlots != minSlotsInt64 {
		proj, err = s.admin.DB.UpdateProject(r.Context(), proj.ID, &database.UpdateProjectOptions{
			Name: proj.Name, Description: proj.Description, Public: proj.Public,
			DirectoryName: proj.DirectoryName, ArchiveAssetID: proj.ArchiveAssetID,
			GitRemote: proj.GitRemote, GithubInstallationID: proj.GithubInstallationID,
			GithubRepoID: proj.GithubRepoID, ManagedGitRepoID: proj.ManagedGitRepoID,
			Subpath: proj.Subpath, ProdVersion: proj.ProdVersion, PrimaryBranch: proj.PrimaryBranch,
			PrimaryDeploymentID: proj.PrimaryDeploymentID, ProdSlots: proj.ProdSlots,
			ProdTTLSeconds: proj.ProdTTLSeconds, DevSlots: proj.DevSlots, DevTTLSeconds: proj.DevTTLSeconds,
			Provisioner: proj.Provisioner, Annotations: proj.Annotations,
			ChcClusterSize: &info.MaxMemoryGB, RillMinSlots: &minSlotsInt64,
			InfraSlots: proj.InfraSlots,
		})
		if err != nil {
			http.Error(w, "failed to update cluster info", http.StatusInternalServerError)
			return nil
		}
	}

	// Auto-scale based on status
	const autoScaleAnnotation = "rill.dev/chc-auto-scaled-slots"
	action := "none"
	if info.Status == "idle" || info.Status == "stopped" {
		if proj.ProdSlots > 1 {
			annotations := maps.Clone(proj.Annotations)
			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations[autoScaleAnnotation] = "true"
			updated, updateErr := s.admin.UpdateProject(r.Context(), proj, &database.UpdateProjectOptions{
				Name: proj.Name, Description: proj.Description, Public: proj.Public,
				DirectoryName: proj.DirectoryName, ArchiveAssetID: proj.ArchiveAssetID,
				GitRemote: proj.GitRemote, GithubInstallationID: proj.GithubInstallationID,
				GithubRepoID: proj.GithubRepoID, ManagedGitRepoID: proj.ManagedGitRepoID,
				Subpath: proj.Subpath, ProdVersion: proj.ProdVersion, PrimaryBranch: proj.PrimaryBranch,
				PrimaryDeploymentID: proj.PrimaryDeploymentID, ProdSlots: 1,
				ProdTTLSeconds: proj.ProdTTLSeconds, DevSlots: proj.DevSlots, DevTTLSeconds: proj.DevTTLSeconds,
				Provisioner: proj.Provisioner, Annotations: annotations,
				ChcClusterSize: proj.ChcClusterSize, RillMinSlots: proj.RillMinSlots,
				InfraSlots: proj.InfraSlots,
			})
			if updateErr == nil {
				proj = updated
				action = "scaled_down"
			}
		}
	} else if info.Status == "running" {
		if proj.RillMinSlots != nil && proj.ProdSlots < int(*proj.RillMinSlots) {
			annotations := maps.Clone(proj.Annotations)
			if annotations != nil {
				delete(annotations, autoScaleAnnotation)
			}
			updated, updateErr := s.admin.UpdateProject(r.Context(), proj, &database.UpdateProjectOptions{
				Name: proj.Name, Description: proj.Description, Public: proj.Public,
				DirectoryName: proj.DirectoryName, ArchiveAssetID: proj.ArchiveAssetID,
				GitRemote: proj.GitRemote, GithubInstallationID: proj.GithubInstallationID,
				GithubRepoID: proj.GithubRepoID, ManagedGitRepoID: proj.ManagedGitRepoID,
				Subpath: proj.Subpath, ProdVersion: proj.ProdVersion, PrimaryBranch: proj.PrimaryBranch,
				PrimaryDeploymentID: proj.PrimaryDeploymentID, ProdSlots: int(*proj.RillMinSlots),
				ProdTTLSeconds: proj.ProdTTLSeconds, DevSlots: proj.DevSlots, DevTTLSeconds: proj.DevTTLSeconds,
				Provisioner: proj.Provisioner, Annotations: annotations,
				ChcClusterSize: proj.ChcClusterSize, RillMinSlots: proj.RillMinSlots,
				InfraSlots: proj.InfraSlots,
			})
			if updateErr == nil {
				proj = updated
				action = "restored"
			}
		}
	}

	result := map[string]interface{}{
		"cloud_status":        info.Status,
		"cloud_service_name":  info.Name,
		"cloud_provider":      info.CloudProvider,
		"cloud_region":        info.Region,
		"cloud_tier":          info.Tier,
		"cloud_min_memory_gb": info.MinMemoryGB,
		"cloud_max_memory_gb": info.MaxMemoryGB,
		"cloud_num_replicas":  info.NumReplicas,
		"prod_slots":          proj.ProdSlots,
		"action":              action,
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(result)
}
