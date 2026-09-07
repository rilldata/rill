package project

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/version"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestEditCloudEditingUsesDedicatedUpdate(t *testing.T) {
	for _, disabled := range []bool{true, false} {
		flag := "--cloud-editing-disabled=true"
		if !disabled {
			flag = "--cloud-editing-disabled=false"
		}
		t.Run(flag, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			server := grpc.NewServer()
			requests := make(chan *adminv1.SudoUpdateProjectCloudEditingRequest, 1)
			// Every RPC except the dedicated update is unimplemented, including GetProject
			// and the full-map replacement API.
			adminv1.RegisterAdminServiceServer(server, &cloudEditingServer{requests: requests})
			t.Cleanup(server.Stop)
			go func() { _ = server.Serve(listener) }()

			ch, err := cmdutil.NewHelper(version.Version{}, t.TempDir())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, ch.Close()) })
			ch.AdminURLOverride = "http://" + listener.Addr().String()
			ch.Printer.OverrideHumanOutput(io.Discard)
			ch.Printer.OverrideDataOutput(io.Discard)
			cmd := EditCmd(ch)
			cmd.SetArgs([]string{"org", "project", flag})
			require.NoError(t, cmd.ExecuteContext(t.Context()))

			req := <-requests
			require.Equal(t, "org", req.Org)
			require.Equal(t, "project", req.Project)
			require.Equal(t, disabled, req.Disabled)
		})
	}
}

type cloudEditingServer struct {
	adminv1.UnimplementedAdminServiceServer
	requests chan *adminv1.SudoUpdateProjectCloudEditingRequest
}

func (s *cloudEditingServer) SudoUpdateProjectCloudEditing(_ context.Context, req *adminv1.SudoUpdateProjectCloudEditingRequest) (*adminv1.SudoUpdateProjectCloudEditingResponse, error) {
	s.requests <- req
	return &adminv1.SudoUpdateProjectCloudEditingResponse{Project: &adminv1.Project{Name: req.Project, OrgName: req.Org}}, nil
}
