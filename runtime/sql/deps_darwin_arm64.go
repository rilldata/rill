//go:build darwin && arm64

package sql

import "embed"

//go:embed deps/darwin_arm64/librillsql.dylib
var libraryFS embed.FS

var libraryPath = "deps/darwin_arm64/librillsql.dylib"
