package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/dbt_cloud"
	"go.uber.org/zap"
)

const maxWebhookBodySize = 1 << 20 // 1MB

// dbtWebhookPayload represents the payload sent by dbt Cloud webhook events.
type dbtWebhookPayload struct {
	EventID   string `json:"eventId"`
	AccountID int    `json:"accountId"`
	Data      struct {
		EventType string `json:"eventType"` // e.g. "job.run.completed"
		RunID     int    `json:"runId"`
		JobID     int    `json:"jobId"`
		RunStatus string `json:"runStatus"` // "Success", "Error", "Cancelled"
	} `json:"data"`
}

// dbtWebhookHandler handles incoming dbt Cloud webhook events.
// Route: POST /v1/instances/{instance_id}/dbt/{connector}/webhook
func (s *Server) dbtWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	instanceID := r.PathValue("instance_id")
	connectorName := r.PathValue("connector")
	ctx := r.Context()

	if instanceID == "" || connectorName == "" {
		http.Error(w, "instance_id and connector are required", http.StatusBadRequest)
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBodySize))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Acquire the connector handle
	handle, release, err := s.runtime.AcquireHandle(ctx, instanceID, connectorName)
	if err != nil {
		s.logger.Warn("dbt webhook: connector not found", zap.String("connector", connectorName), zap.Error(err))
		http.Error(w, "connector not found", http.StatusNotFound)
		return
	}
	defer release()

	if handle.Driver() != "dbt_cloud" {
		http.Error(w, "not a dbt_cloud connector", http.StatusBadRequest)
		return
	}

	// Validate HMAC signature if webhook_secret is configured
	webhookSecret, _ := handle.Config()["webhook_secret"].(string)
	if webhookSecret != "" {
		signature := r.Header.Get("Authorization")
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse the webhook payload
	var payload dbtWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		s.logger.Warn("dbt webhook: failed to parse payload", zap.Error(err))
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	s.logger.Info("dbt webhook received",
		zap.String("event_type", payload.Data.EventType),
		zap.String("run_status", payload.Data.RunStatus),
		zap.Int("run_id", payload.Data.RunID),
	)

	// Only process job.run.completed events
	if payload.Data.EventType != "job.run.completed" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Invalidate the cached manifest
	if invalidator, ok := handle.(dbt_cloud.ManifestInvalidator); ok {
		invalidator.InvalidateManifest()
	}

	if payload.Data.RunStatus == "Success" {
		// Trigger refresh for all models that use this dbt_cloud connector as input
		if err := s.triggerDbtModelRefresh(ctx, instanceID, connectorName); err != nil {
			s.logger.Error("dbt webhook: failed to trigger model refresh", zap.Error(err))
			http.Error(w, "failed to trigger refresh", http.StatusInternalServerError)
			return
		}
	} else {
		s.logger.Warn("dbt webhook: run did not succeed",
			zap.String("run_status", payload.Data.RunStatus),
			zap.Int("run_id", payload.Data.RunID),
		)
		// On failure, the models remain with their last successfully synced data.
		// The next reconciliation will surface a warning if the manifest cannot be fetched.
	}

	w.WriteHeader(http.StatusOK)
}

// triggerDbtModelRefresh creates a refresh trigger for all models that depend on the given dbt_cloud connector.
func (s *Server) triggerDbtModelRefresh(ctx context.Context, instanceID, connectorName string) error {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return err
	}

	// List all model resources
	resources, err := ctrl.List(ctx, runtime.ResourceKindModel, "", false)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Find models whose input connector matches the dbt_cloud connector
	var modelTriggers []*runtimev1.RefreshModelTrigger
	for _, r := range resources {
		model := r.GetModel()
		if model == nil {
			continue
		}
		if model.Spec.InputConnector == connectorName {
			modelTriggers = append(modelTriggers, &runtimev1.RefreshModelTrigger{
				Model: r.Meta.Name.Name,
				Full:  true,
			})
		}
	}

	if len(modelTriggers) == 0 {
		return nil
	}

	// Create a refresh trigger resource
	spec := &runtimev1.RefreshTriggerSpec{
		Models: modelTriggers,
	}
	name := fmt.Sprintf("dbt_webhook_%s", randomString(8))
	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: name}
	res := &runtimev1.Resource{Resource: &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{Spec: spec}}}
	if err := ctrl.Create(ctx, n, nil, nil, nil, false, res); err != nil {
		return fmt.Errorf("failed to create refresh trigger: %w", err)
	}

	return nil
}
