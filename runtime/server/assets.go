package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) assetsHandler(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	path := req.PathValue("path")

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
		attribute.String("args.path", path),
	)

	if !auth.GetClaims(req.Context(), instanceID).Can(runtime.ReadObjects) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to assets")
	}

	inst, err := s.runtime.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	allowed := false
	for _, p := range inst.PublicPaths {
		// 'p' can be `/public`, `/public/`, `public/`, `public` (with os-based separators)
		// match pattern `public/*` or `/public/*`
		ok, err := filepath.Match(fmt.Sprintf("%s%c*", filepath.Clean(p), os.PathSeparator), path)
		if err != nil {
			return httputil.Error(http.StatusBadRequest, err)
		}
		if ok {
			allowed = true
			break
		}
	}
	if !allowed {
		return httputil.Error(http.StatusForbidden, fmt.Errorf("path is not allowed"))
	}

	repo, release, err := s.runtime.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	str, err := repo.Get(ctx, path)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(str))
	return err
}
