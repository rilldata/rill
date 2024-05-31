//go:build linux && amd64

package extensions

import "embed"

//go:embed embed/linux_amd64/*
var embeddedFiles embed.FS
