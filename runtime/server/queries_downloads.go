package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries/downloads"
)

func (s *Server) Download(ctx context.Context, req *runtimev1.DownloadLinkRequest) (*runtimev1.DownloadLinkResponse, error) {
	r, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf("/v1/downloads?%s=%s&%s=%d&%s=%s&%s=%s", downloads.Request, base64.StdEncoding.EncodeToString(r), downloads.Limit, req.Limit, downloads.Format, req.Format, downloads.Compression, req.Compression)

	return &runtimev1.DownloadLinkResponse{
		DownloadUrlPath: out,
	}, nil
}
