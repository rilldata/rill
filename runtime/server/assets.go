package server

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// assetContentSecurityPolicy is the Content-Security-Policy of project assets.
// It only applies when a browser opens an asset as a document, which it then renders in a sandbox without script.
// It has no effect on assets loaded as subresources, such as images, fonts or fetch requests.
const assetContentSecurityPolicy = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

// assetContentTypes maps the file extensions of project assets to their content types.
// Assets with other extensions are served as application/octet-stream.
// The type is never sniffed from the content: browsers only render some files when the type is right (notably SVG),
// and a sniffed type would let a file's content decide how the browser treats it.
var assetContentTypes = map[string]string{
	".avif":  "image/avif",
	".bmp":   "image/bmp",
	".gif":   "image/gif",
	".ico":   "image/x-icon",
	".jpeg":  "image/jpeg",
	".jpg":   "image/jpeg",
	".png":   "image/png",
	".svg":   "image/svg+xml",
	".webp":  "image/webp",
	".otf":   "font/otf",
	".ttf":   "font/ttf",
	".woff":  "font/woff",
	".woff2": "font/woff2",
	".css":   "text/css; charset=utf-8",
	".csv":   "text/csv; charset=utf-8",
	".json":  "application/json",
	// Map boundaries.
	".geojson":  "application/json",
	".topojson": "application/json",
	".tsv":      "text/tab-separated-values; charset=utf-8",
	".txt":      "text/plain; charset=utf-8",
}

// assetsHandler serves files from the project's public paths (see `public_paths` in rill.yaml), including files in subdirectories.
func (s *Server) assetsHandler(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
		attribute.String("args.path", req.PathValue("path")),
	)

	if !auth.GetClaims(req.Context(), instanceID).Can(runtime.ReadObjects) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to assets")
	}

	inst, err := s.runtime.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	// Resolve the requested path to a rooted, slash-separated path without `.` or `..` segments before matching it.
	// This keeps a request like `public/../rill.yaml` from reading files outside the public paths.
	// The handler can't leave it to the ServeMux: it passes on `..` segments that were percent-encoded in the URL.
	assetPath := path.Clean("/" + filepath.ToSlash(req.PathValue("path")))

	allowed := false
	for _, p := range inst.PublicPaths {
		// A public path may be written as `public`, `/public`, `./public` or `public/` (with OS-specific separators).
		p = path.Clean("/" + filepath.ToSlash(p))
		if p == "/" {
			// Never expose the entire project.
			continue
		}
		if assetPath == p || strings.HasPrefix(assetPath, p+"/") {
			allowed = true
			break
		}
	}
	if !allowed {
		return httputil.Errorf(http.StatusForbidden, "path is not allowed")
	}

	repo, release, err := s.runtime.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	data, err := repo.Get(ctx, assetPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return httputil.Errorf(http.StatusNotFound, "file %q not found", assetPath)
		}
		return err
	}

	contentType, ok := assetContentTypes[strings.ToLower(path.Ext(assetPath))]
	if !ok {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", assetContentSecurityPolicy)

	// Assets may hold licensed data, such as map boundaries, so shared caches must never store them.
	// Browsers revalidate on every use, which is cheap thanks to the ETag.
	sum := sha256.Sum256([]byte(data))
	w.Header().Set("Cache-Control", "private, no-cache")
	w.Header().Set("ETag", `"`+hex.EncodeToString(sum[:])+`"`)

	// ServeContent answers conditional and range requests; it keeps the Content-Type set above.
	http.ServeContent(w, req, assetPath, time.Time{}, strings.NewReader(data))
	return nil
}
