package server

import (
	"context"
	"encoding/base64"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
)

func (s *Server) Download(ctx context.Context, req *runtimev1.DownloadLinkRequest) (*runtimev1.DownloadLinkResponse, error) {
	r, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf("/v1/downloads?%s=%s", "request", base64.StdEncoding.EncodeToString(r))

	return &runtimev1.DownloadLinkResponse{
		DownloadUrlPath: out,
	}, nil
}
