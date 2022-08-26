//go:build darwin && amd64

package sql

import "embed"

//go:embed deps/darwin_amd64/librillsql.dylib
var libraryFS embed.FS

var libraryPath = "deps/darwin_amd64/librillsql.dylib"
