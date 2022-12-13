package web

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
)

//go:embed all:embed
var distFS embed.FS

// Handler serves an web-local UI
func StaticHandler() http.Handler {
	uiAssetFS := newUIAssetFS()
	return gziphandler.GzipHandler(http.FileServer(uiAssetFS))
}

// Check if web-local dist static UI is exists, If not server the default index.html page
func newUIAssetFS() http.FileSystem {
	_, err := distFS.ReadFile("embed/dist/index.html")
	if os.IsNotExist(err) {
		return assetFS(distFS, "embed")
	}
	return assetFS(distFS, "embed/dist")
}

// Get the subtree of the embedded files with `embed` directory as a root.
func assetFS(embeddedFS embed.FS, dir string) http.FileSystem {
	subFS, err := fs.Sub(embeddedFS, dir)
	if err != nil {
		panic(fmt.Errorf("fs embed: %w", err))
	}

	return &SPARoutingFS{FileSystem: http.FS(subFS)}
}

type SPARoutingFS struct {
	FileSystem http.FileSystem
}

func (spaFS *SPARoutingFS) Open(name string) (http.File, error) {
	file, err := spaFS.FileSystem.Open(name)
	if err == nil {
		return file, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		file, err = spaFS.FileSystem.Open("index.html")
		return file, err
	}

	return nil, err
}
