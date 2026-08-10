package local

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	localv1 "github.com/rilldata/rill/proto/gen/rill/local/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCloudFeedbackErrState(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantState localv1.CloudFeedbackState
		wantInMsg string
	}{
		{
			name:      "permission denied",
			err:       status.Error(codes.PermissionDenied, "forbidden"),
			wantState: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NO_PERMISSION,
			wantInMsg: "admin access",
		},
		{
			name:      "unimplemented on outdated runtime",
			err:       status.Error(codes.Unimplemented, "unknown method"),
			wantState: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR,
			wantInMsg: "latest Rill version",
		},
		{
			name:      "generic error",
			err:       errors.New("connection refused"),
			wantState: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR,
			wantInMsg: "connection refused",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, msg := cloudFeedbackErrState(tc.err)
			require.Equal(t, tc.wantState, state)
			require.Contains(t, msg, tc.wantInMsg)
		})
	}
}

func TestCloudRuntimeResultConnectErr(t *testing.T) {
	// OK state produces no error.
	res := &cloudRuntimeResult{state: localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_OK}
	require.NoError(t, res.connectErr())

	// A NotFound runtime error is preserved as a NotFound connect error.
	res = &cloudRuntimeResult{
		state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_ERROR,
		stateMsg: "feedback not found",
		err:      status.Error(codes.NotFound, "feedback not found"),
	}
	err := res.connectErr()
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))

	// Precondition states map to FailedPrecondition.
	res = &cloudRuntimeResult{
		state:    localv1.CloudFeedbackState_CLOUD_FEEDBACK_STATE_NOT_LOGGED_IN,
		stateMsg: "Not logged in to Rill Cloud",
	}
	err = res.connectErr()
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "Not logged in")
}
