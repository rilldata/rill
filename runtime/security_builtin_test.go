package runtime

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBuiltInReportAuthorizationMatrix(t *testing.T) {
	// Check each identity source that may grant access to an otherwise private scheduled report.
	emailProps, err := structpb.NewStruct(map[string]any{"recipients": []any{"recipient@example.com"}})
	require.NoError(t, err)
	slackProps, err := structpb.NewStruct(map[string]any{"users": []any{"slack@example.com"}})
	require.NoError(t, err)

	// This matrix documents every identity source that can grant access to an
	// otherwise private scheduled report.
	tests := []struct {
		name    string
		attrs   map[string]any
		allowed bool
	}{
		{name: "admin", attrs: map[string]any{"admin": true}, allowed: true},
		{name: "owner", attrs: map[string]any{"id": "owner-id"}, allowed: true},
		{name: "email recipient", attrs: map[string]any{"email": "recipient@example.com"}, allowed: true},
		{name: "Slack user", attrs: map[string]any{"email": "slack@example.com"}, allowed: true},
		{name: "unrelated user", attrs: map[string]any{"id": "other", "email": "other@example.com"}, allowed: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := newSecurityTestReport("report", map[string]string{"admin_owner_user_id": "owner-id"}, []*runtimev1.Notifier{
				{Connector: "email", Properties: emailProps},
				{Connector: "slack", Properties: slackProps},
			})
			security, err := newSecurityEngine(10, zap.NewNop(), nil).resolveSecurity(t.Context(), "instance", "prod", nil, &SecurityClaims{UserAttributes: tt.attrs}, res)
			require.NoError(t, err)
			require.Equal(t, tt.allowed, security.CanAccess())
		})
	}
}

func TestBuiltInAlertAuthorizationMatrix(t *testing.T) {
	// Check owner, recipient, Slack-user, channel, and admin access through the public security result.
	emailProps, err := structpb.NewStruct(map[string]any{"recipients": []any{"recipient@example.com"}})
	require.NoError(t, err)
	slackProps, err := structpb.NewStruct(map[string]any{
		"users":    []any{"slack@example.com"},
		"channels": []any{"alerts"},
	})
	require.NoError(t, err)

	// Channel-only access is intentionally characterized here because the
	// current policy grants every project user access when any channel is set.
	tests := []struct {
		name    string
		attrs   map[string]any
		allowed bool
	}{
		{name: "admin", attrs: map[string]any{"admin": true}, allowed: true},
		{name: "owner", attrs: map[string]any{"id": "owner-id"}, allowed: true},
		{name: "email recipient", attrs: map[string]any{"email": "recipient@example.com"}, allowed: true},
		{name: "Slack user", attrs: map[string]any{"email": "slack@example.com"}, allowed: true},
		{name: "Slack channel project user", attrs: map[string]any{"email": "unrelated@example.com"}, allowed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := newSecurityTestAlert("alert", map[string]string{"admin_owner_user_id": "owner-id"}, []*runtimev1.Notifier{
				{Connector: "email", Properties: emailProps},
				{Connector: "slack", Properties: slackProps},
			})
			security, err := newSecurityEngine(10, zap.NewNop(), nil).resolveSecurity(t.Context(), "instance", "prod", nil, &SecurityClaims{UserAttributes: tt.attrs}, res)
			require.NoError(t, err)
			require.Equal(t, tt.allowed, security.CanAccess())
		})
	}
}

func TestBuiltInCanvasAuthorizationMatrix(t *testing.T) {
	// Admin-managed canvases are private unless owned or explicitly shared;
	// ordinary canvases remain accessible by default.
	tests := []struct {
		name        string
		annotations map[string]string
		attrs       map[string]any
		allowed     bool
	}{
		{name: "personal owner", annotations: map[string]string{"admin_managed": "true", "admin_owner_user_id": "owner-id"}, attrs: map[string]any{"id": "owner-id"}, allowed: true},
		{name: "personal admin", annotations: map[string]string{"admin_managed": "true"}, attrs: map[string]any{"admin": true}, allowed: true},
		{name: "personal unrelated", annotations: map[string]string{"admin_managed": "true", "admin_owner_user_id": "owner-id"}, attrs: map[string]any{"id": "other"}, allowed: false},
		{name: "shared personal", annotations: map[string]string{"admin_managed": "true", "admin_shared": "true"}, attrs: map[string]any{"id": "other"}, allowed: true},
		{name: "ordinary", annotations: nil, attrs: map[string]any{"id": "other"}, allowed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := newSecurityTestCanvas("canvas", tt.annotations)
			security, err := newSecurityEngine(10, zap.NewNop(), nil).resolveSecurity(t.Context(), "instance", "prod", nil, &SecurityClaims{UserAttributes: tt.attrs}, res)
			require.NoError(t, err)
			require.Equal(t, tt.allowed, security.CanAccess())
		})
	}
}

func TestBuiltInAuthorizationMalformedResourcesFailClosed(t *testing.T) {
	// Partially reconciled resources are observable during failures. Security
	// resolution must deny them instead of panicking or defaulting open.
	tests := []*runtimev1.Resource{
		{Meta: newSecurityTestMeta(ResourceKindReport, "nil-report"), Resource: &runtimev1.Resource_Report{Report: &runtimev1.Report{}}},
		{Meta: newSecurityTestMeta(ResourceKindAlert, "nil-alert"), Resource: &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{}}},
		{Meta: newSecurityTestMeta(ResourceKindCanvas, "nil-canvas"), Resource: &runtimev1.Resource_Canvas{Canvas: &runtimev1.Canvas{}}},
		{Meta: newSecurityTestMeta(ResourceKindMetricsView, "nil-metrics"), Resource: &runtimev1.Resource_MetricsView{MetricsView: &runtimev1.MetricsView{}}},
		{Meta: newSecurityTestMeta(ResourceKindAPI, "nil-api"), Resource: &runtimev1.Resource_Api{Api: &runtimev1.API{}}},
	}
	for _, res := range tests {
		t.Run(res.Meta.Name.Name, func(t *testing.T) {
			security, err := newSecurityEngine(10, zap.NewNop(), nil).resolveSecurity(t.Context(), "instance", "prod", nil, &SecurityClaims{UserAttributes: map[string]any{"admin": true}}, res)
			require.NoError(t, err)
			require.False(t, security.CanAccess())
		})
	}
}

func newSecurityTestReport(name string, annotations map[string]string, notifiers []*runtimev1.Notifier) *runtimev1.Resource {
	return &runtimev1.Resource{
		Meta: newSecurityTestMeta(ResourceKindReport, name),
		Resource: &runtimev1.Resource_Report{Report: &runtimev1.Report{Spec: &runtimev1.ReportSpec{
			Annotations: annotations,
			Notifiers:   notifiers,
		}}},
	}
}

func newSecurityTestAlert(name string, annotations map[string]string, notifiers []*runtimev1.Notifier) *runtimev1.Resource {
	return &runtimev1.Resource{
		Meta: newSecurityTestMeta(ResourceKindAlert, name),
		Resource: &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{Spec: &runtimev1.AlertSpec{
			Annotations: annotations,
			Notifiers:   notifiers,
		}}},
	}
}

func newSecurityTestCanvas(name string, annotations map[string]string) *runtimev1.Resource {
	spec := &runtimev1.CanvasSpec{Annotations: annotations}
	return &runtimev1.Resource{
		Meta: newSecurityTestMeta(ResourceKindCanvas, name),
		Resource: &runtimev1.Resource_Canvas{Canvas: &runtimev1.Canvas{
			Spec:  spec,
			State: &runtimev1.CanvasState{ValidSpec: spec},
		}},
	}
}

func newSecurityTestMeta(kind, name string) *runtimev1.ResourceMeta {
	return &runtimev1.ResourceMeta{
		Name:           &runtimev1.ResourceName{Kind: kind, Name: name},
		StateUpdatedOn: timestamppb.New(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)),
	}
}
