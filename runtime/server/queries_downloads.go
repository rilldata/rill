package server

import (
	"context"
	"encoding/base64"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries/downloads"
	"google.golang.org/protobuf/proto"
)

func (s *Server) Download(ctx context.Context, req *runtimev1.DownloadLinkRequest) (*runtimev1.DownloadLinkResponse, error) {
	r, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf("/v1/downloads?%s=%s", downloads.Request, base64.StdEncoding.EncodeToString(r))

	if req.Limit > 0 {
		out += fmt.Sprintf("&%s=%d", downloads.Limit, req.Limit)
	}

	if req.Format != runtimev1.DownloadFormat_DOWNLOAD_FORMAT_UNSPECIFIED {
		out += fmt.Sprintf("&%s=%s", downloads.Format, req.Format)
	}

	if req.Compression != runtimev1.DownloadCompression_DOWNLOAD_COMPRESSION_UNSPECIFIED {
		out += fmt.Sprintf("&%s=%s", downloads.Compression, req.Compression)
	}

	return &runtimev1.DownloadLinkResponse{
		DownloadUrlPath: out,
	}, nil
}
