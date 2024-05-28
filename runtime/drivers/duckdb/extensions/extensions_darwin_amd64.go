//go:build darwin && amd64

package extensions

import "embed"

//go:embed embed/osx_amd64/*
var embeddedFiles embed.FS
