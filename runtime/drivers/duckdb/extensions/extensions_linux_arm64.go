//go:build linux && arm64

package extensions

import "embed"

//go:embed embed/linux_arm64/*
var embeddedFiles embed.FS
