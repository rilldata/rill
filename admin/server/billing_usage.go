package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/rilldata/rill/admin/database"
	"go.opentelemetry.io/otel/attribute"
)

// recordExternalUserUsage emits the billable external_users metric for an embedded external user (one
// identified by an external_user_id). It is emitted once per embed token issuance rather than per runtime
// request: the downstream billing metric counts distinct users per period, so issuance is a sufficient and
// far cheaper signal. The user is identified by the same hashed, non-PII subject used in the runtime JWT.
func (s *Server) recordExternalUserUsage(ctx context.Context, proj *database.Project, depl *database.Deployment, externalUserID string) {
	s.recordEmbeddedUserUsage(ctx, "external_users", proj, depl, attribute.String("user_id", subjectForExternalUser(externalUserID, proj.ID)))
}

// recordAnonymousEmbedUsage emits the billable external_anonymous_users metric for an embedded request with
// no external_user_id. Distinct counting downstream uses a non-PII hash of the user attributes.
func (s *Server) recordAnonymousEmbedUsage(ctx context.Context, proj *database.Project, depl *database.Deployment, attrs map[string]any) {
	s.recordEmbeddedUserUsage(ctx, "external_anonymous_users", proj, depl, attribute.String("external_anonymous_user", anonymousUserID(attrs)))
}

// recordEmbeddedUserUsage emits an embedded-user usage metric on the billing events pipeline, attributed to the
// organization and project (matching the annotation keys the runtime emits, so the events aggregate together).
func (s *Server) recordEmbeddedUserUsage(ctx context.Context, metricName string, proj *database.Project, depl *database.Deployment, idAttr attribute.KeyValue) {
	s.billingActivity.RecordMetric(ctx, metricName, 1,
		attribute.String("organization_id", proj.OrganizationID),
		attribute.String("project_id", proj.ID),
		attribute.String("project_name", proj.Name),
		attribute.String("instance_id", depl.RuntimeInstanceID),
		idAttr,
	)
}

// anonymousUserID returns a deterministic, non-PII identifier for an anonymous embedded user, derived from
// their user attributes. Downstream billing counts distinct values of this identifier.
func anonymousUserID(attrs map[string]any) string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%v\x00", k, attrs[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}
