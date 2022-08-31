//go:build linux && amd64

package sql

import "embed"

//go:embed deps/linux_amd64/librillsql.so
var libraryFS embed.FS

var libraryPath = "deps/linux_amd64/librillsql.so"
