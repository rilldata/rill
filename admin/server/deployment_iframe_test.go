package server

import (
	"strings"
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/embedopen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestGetIFrameValidation tests the validation of GetIFrame's open and external_user_id params, which happens before any database access.
func TestGetIFrameValidation(t *testing.T) {
	open, err := embedopen.Encode(&embedopen.Payload{Kind: embedopen.KindReport, Name: "r1"})
	require.NoError(t, err)
	openPath, err := embedopen.Encode(&embedopen.Payload{Kind: embedopen.KindReport, Name: "../personal/c1"})
	require.NoError(t, err)
	openBackslash, err := embedopen.Encode(&embedopen.Payload{Kind: embedopen.KindReport, Name: `..\personal\c1`})
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
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Open: "not-valid", ExternalUserId: "u1"},
			wantErr: "invalid open",
		},
		{
			name:    "open with a path as name",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Open: openPath, ExternalUserId: "u1"},
			wantErr: "name can't contain slashes",
		},
		{
			name:    "open with a backslash in name",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Open: openBackslash, ExternalUserId: "u1"},
			wantErr: "name can't contain slashes",
		},
		{
			name:    "open without external user ID",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Open: open},
			wantErr: "open requires external_user_id",
		},
		{
			name:    "no resource without navigation",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p"},
			wantErr: "resource must be provided",
		},
		{
			name:    "external user ID too long",
			req:     &adminv1.GetIFrameRequest{Org: "o", Project: "p", Resource: "e1", ExternalUserId: strings.Repeat("u", maxExternalUserIDBytes+1)},
			wantErr: "external_user_id must be at most 2048 bytes",
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
