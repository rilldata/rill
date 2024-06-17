//go:build darwin && arm64

package extensions

import "embed"

//go:embed embed/osx_arm64/*
var embeddedFiles embed.FS
