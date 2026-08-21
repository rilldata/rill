package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// cloudRuntimeTokenTTL is how long a cached cloud runtime client is reused.
// The JWT returned by AdminService.GetProject is valid for 30 minutes; refreshing at 20 leaves a safe margin.
const cloudRuntimeTokenTTL = 20 * time.Minute

// cloudRuntimeCloseGrace is how long a replaced cloud runtime client lingers before being closed,
// so requests that acquired it just before the replacement can finish.
const cloudRuntimeCloseGrace = 2 * time.Minute

// cloudRuntimeCache caches an authenticated runtime client for the current project's cloud deployment.
// It is keyed on the admin token in addition to the project, so switching accounts in the running
// local app can never serve a client that carries the previous user's JWT.
type cloudRuntimeCache struct {
	mu           sync.Mutex
	adminTokenID string
	org          string
	project      string
	instanceID   string
	client       *runtimeclient.Client
	expiresOn    time.Time
}

// cloudRuntimeResult is the outcome of acquiring (and optionally calling) the cloud deployment's runtime.
// A non-OK state describes why the cloud feedback store is unreachable; err holds the underlying call error, if any.
type cloudRuntimeResult struct {
	client     *runtimeclient.Client
	instanceID string
	org        string
	project    string
	state      localv1.CloudFeedbackState
	stateMsg   string
	err        error
}

func (s *Server) ListProjectAIFeedback(ctx context.Context, r *connect.Request[localv1.ListProjectAIFeedbackRequest]) (*connect.Response[localv1.ListProjectAIFeedbackResponse], error) {
	feedbackStatus := r.Msg.Status
	if feedbackStatus == "" {
		feedbackStatus = "open"
	}

	var out *runtimev1.ListAIFeedbackResponse
	res := s.withCloudRuntime(ctx, func(rt *runtimeclient.Client, instanceID string) error {
		var err error
		out, err = rt.ListAIFeedback(ctx, &runtimev1.ListAIFeedbackRequest{
			InstanceId: instanceID,
			Status:     feedbackStatus,
			Kind:       r.Msg.Kind,
			PageSize:   r.Msg.PageSize,
			PageToken:  r.Msg.PageToken,
		})
		return err
	})

	resp := &localv1.ListProjectAIFeedbackResponse{
		CloudState:        res.state,
		CloudStateMessage: res.stateMsg,
		Org:               res.org,
		Project:           res.project,
	}
	if res.state == localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_OK {
		resp.Feedback = out.Feedback
		resp.NextPageToken = out.NextPageToken
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) GetProjectAIFeedback(ctx context.Context, r *connect.Request[localv1.GetProjectAIFeedbackRequest]) (*connect.Response[localv1.GetProjectAIFeedbackResponse], error) {
	var out *runtimev1.GetAIFeedbackResponse
	res := s.withCloudRuntime(ctx, func(rt *runtimeclient.Client, instanceID string) error {
		var err error
		out, err = rt.GetAIFeedback(ctx, &runtimev1.GetAIFeedbackRequest{
			InstanceId: instanceID,
			FeedbackId: r.Msg.FeedbackId,
		})
		return err
	})
	if err := res.connectErr(); err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.GetProjectAIFeedbackResponse{
		Feedback:     out.Feedback,
		Conversation: out.Conversation,
		Messages:     out.Messages,
	}), nil
}

func (s *Server) ResolveProjectAIFeedback(ctx context.Context, r *connect.Request[localv1.ResolveProjectAIFeedbackRequest]) (*connect.Response[localv1.ResolveProjectAIFeedbackResponse], error) {
	var out *runtimev1.UpdateAIFeedbackStatusResponse
	res := s.withCloudRuntime(ctx, func(rt *runtimeclient.Client, instanceID string) error {
		var err error
		out, err = rt.UpdateAIFeedbackStatus(ctx, &runtimev1.UpdateAIFeedbackStatusRequest{
			InstanceId: instanceID,
			FeedbackId: r.Msg.FeedbackId,
			Status:     r.Msg.Status,
		})
		return err
	})
	if err := res.connectErr(); err != nil {
		return nil, err
	}

	return connect.NewResponse(&localv1.ResolveProjectAIFeedbackResponse{
		Feedback: out.Feedback,
	}), nil
}

// withCloudRuntime acquires a runtime client for the project's cloud deployment and invokes fn with it.
// If fn fails because the cached JWT expired, the client is refreshed and fn retried once.
// Failures are reported as a state on the returned result, never as a Go error,
// so list-style callers can render them as UI states.
func (s *Server) withCloudRuntime(ctx context.Context, fn func(rt *runtimeclient.Client, instanceID string) error) *cloudRuntimeResult {
	res := s.acquireCloudRuntime(ctx)
	if res.state != localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_OK {
		return res
	}

	err := fn(res.client, res.instanceID)
	if err != nil && status.Code(err) == codes.Unauthenticated {
		s.invalidateCloudRuntime()
		res = s.acquireCloudRuntime(ctx)
		if res.state != localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_OK {
			return res
		}
		err = fn(res.client, res.instanceID)
	}
	if err != nil {
		res.err = err
		res.state, res.stateMsg = cloudFeedbackErrState(err)
	}
	return res
}

// acquireCloudRuntime resolves the current project's cloud deployment and returns an authenticated runtime client for it.
// Preconditions (logged in, project deployed, admin permissions) are reported as states rather than errors.
// Network calls run outside the cache lock, and replaced clients are closed after a grace period,
// so concurrent feedback requests neither serialize behind the admin service nor lose their connection mid-flight.
func (s *Server) acquireCloudRuntime(ctx context.Context) *cloudRuntimeResult {
	if !s.app.ch.IsAuthenticated() {
		return &cloudRuntimeResult{
			state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NOT_LOGGED_IN,
			stateMsg: "Not logged in to Rill Cloud",
		}
	}
	adminTokenID := hashAdminToken(s.app.ch.AdminToken())

	projects, err := s.app.ch.InferProjects(ctx, s.app.ch.Org, s.app.ProjectPath)
	if err != nil {
		if errors.Is(err, cmdutil.ErrInferProjectFailed) {
			return &cloudRuntimeResult{
				state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NOT_DEPLOYED,
				stateMsg: "No matching Rill Cloud project found for this directory",
			}
		}
		return &cloudRuntimeResult{
			state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR,
			stateMsg: err.Error(),
			err:      err,
		}
	}
	proj := projects[0]

	res := &cloudRuntimeResult{
		org:     proj.OrgName,
		project: proj.Name,
		state:   localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_OK,
	}
	if len(projects) > 1 {
		res.stateMsg = fmt.Sprintf("Multiple cloud projects matched this directory; using %s/%s", proj.OrgName, proj.Name)
	}

	// Reuse the cached client if it was minted for the same user and project and its JWT is still fresh.
	s.cloudRuntime.mu.Lock()
	if s.cloudRuntime.client != nil && s.cloudRuntime.adminTokenID == adminTokenID && s.cloudRuntime.org == proj.OrgName && s.cloudRuntime.project == proj.Name && time.Now().Before(s.cloudRuntime.expiresOn) {
		res.client = s.cloudRuntime.client
		res.instanceID = s.cloudRuntime.instanceID
		s.cloudRuntime.mu.Unlock()
		return res
	}
	s.cloudRuntime.mu.Unlock()

	c, err := s.app.ch.Client()
	if err != nil {
		return &cloudRuntimeResult{org: res.org, project: res.project, state: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR, stateMsg: err.Error(), err: err}
	}

	projResp, err := c.GetProject(ctx, &adminv1.GetProjectRequest{Org: proj.OrgName, Project: proj.Name})
	if err != nil {
		return &cloudRuntimeResult{org: res.org, project: res.project, state: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR, stateMsg: err.Error(), err: err}
	}

	// Precheck permissions so the UI gets a precise message instead of a runtime 403.
	// ManageProd mirrors the condition under which the admin service grants the ManageAIFeedback runtime permission.
	if projResp.ProjectPermissions == nil || !projResp.ProjectPermissions.ManageProd {
		return &cloudRuntimeResult{
			org:      res.org,
			project:  res.project,
			state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NO_PERMISSION,
			stateMsg: "You need admin access to the cloud project to review AI feedback",
		}
	}

	if projResp.Deployment == nil || projResp.Deployment.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
		return &cloudRuntimeResult{
			org:      res.org,
			project:  res.project,
			state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NOT_DEPLOYED,
			stateMsg: fmt.Sprintf("The project %q has no running cloud deployment", proj.Name),
		}
	}

	client, err := runtimeclient.New(projResp.Deployment.RuntimeHost, projResp.Jwt)
	if err != nil {
		return &cloudRuntimeResult{org: res.org, project: res.project, state: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR, stateMsg: err.Error(), err: err}
	}

	s.cloudRuntime.mu.Lock()
	// A concurrent request may have refreshed the cache while we were talking to the admin service; prefer its client.
	if s.cloudRuntime.client != nil && s.cloudRuntime.adminTokenID == adminTokenID && s.cloudRuntime.org == proj.OrgName && s.cloudRuntime.project == proj.Name && time.Now().Before(s.cloudRuntime.expiresOn) {
		res.client = s.cloudRuntime.client
		res.instanceID = s.cloudRuntime.instanceID
		s.cloudRuntime.mu.Unlock()
		_ = client.Close() // ours was never handed out
		return res
	}
	closeClientAfterGrace(s.cloudRuntime.client)
	s.cloudRuntime.adminTokenID = adminTokenID
	s.cloudRuntime.org = proj.OrgName
	s.cloudRuntime.project = proj.Name
	s.cloudRuntime.instanceID = projResp.Deployment.RuntimeInstanceId
	s.cloudRuntime.client = client
	s.cloudRuntime.expiresOn = time.Now().Add(cloudRuntimeTokenTTL)
	s.cloudRuntime.mu.Unlock()

	res.client = client
	res.instanceID = projResp.Deployment.RuntimeInstanceId
	return res
}

func (s *Server) invalidateCloudRuntime() {
	s.cloudRuntime.mu.Lock()
	defer s.cloudRuntime.mu.Unlock()
	closeClientAfterGrace(s.cloudRuntime.client)
	s.cloudRuntime.client = nil
}

// closeClientAfterGrace closes a replaced client after a grace period, so in-flight requests
// that acquired it just before the replacement aren't killed mid-RPC.
func closeClientAfterGrace(client *runtimeclient.Client) {
	if client == nil {
		return
	}
	time.AfterFunc(cloudRuntimeCloseGrace, func() { _ = client.Close() })
}

func hashAdminToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// connectErr converts a non-OK result into a connect error.
// Used by RPCs that return a single item, where a state-as-data response isn't meaningful.
func (r *cloudRuntimeResult) connectErr() error {
	if r.state == localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_OK {
		return nil
	}
	if r.err != nil {
		if code := status.Code(r.err); code == codes.NotFound || code == codes.PermissionDenied || code == codes.InvalidArgument {
			return connect.NewError(connect.Code(code), errors.New(r.stateMsg))
		}
	}
	return connect.NewError(connect.CodeFailedPrecondition, errors.New(r.stateMsg))
}

// cloudFeedbackErrState maps a cloud runtime call error to a UI-friendly state.
func cloudFeedbackErrState(err error) (localv1.CloudFeedbackState, string) {
	switch status.Code(err) {
	case codes.PermissionDenied:
		return localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NO_PERMISSION, "You need admin access to the cloud project to review AI feedback"
	case codes.Unimplemented:
		return localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR, "The project's cloud deployment does not support AI feedback review yet. Redeploy it with the latest Rill version."
	default:
		return localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR, err.Error()
	}
}
