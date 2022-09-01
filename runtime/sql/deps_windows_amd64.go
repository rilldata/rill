//go:build windows && amd64

package sql

import "embed"

//go:embed deps/windows_amd64/librillsql.dll
var libraryFS embed.FS

var libraryPath = "deps/windows_amd64/librillsql.dll"
