package server

import (
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/embedopen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestGetIFrameValidation tests the validation of GetIFrame's open and open_url params, which happens before any database access.
func TestGetIFrameValidation(t *testing.T) {
	open, err := embedopen.Encode(&embedopen.Payload{Kind: embedopen.KindReport, Name: "r1"})
	require.NoError(t, err)

	tt := []struct {
		name    string
		req     *adminv1.GetIFrameRequest
		wantErr string
	}{
		{
			name:    "open combined with resource",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Resource: "e1", Open: open},
			wantErr: "open can't be combined with resource",
		},
		{
			name:    "invalid open",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Open: "not-valid"},
			wantErr: "invalid open",
		},
		{
			name:    "relative open URL",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Resource: "e1", OpenUrl: "/open"},
			wantErr: "must be an absolute http or https URL",
		},
		{
			name:    "non-http open URL",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Resource: "e1", OpenUrl: "javascript:alert(1)"},
			wantErr: "must be an absolute http or https URL",
		},
		{
			name:    "no resource without navigation",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p"},
			wantErr: "resource must be provided",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&Server{}).GetIFrame(t.Context(), tc.req)
			require.ErrorContains(t, err, tc.wantErr)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}
