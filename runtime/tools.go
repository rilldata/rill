//go:build tools

package runtime

// Tools installed with go install that `go mod tidy` should keep.
import (
	_ "src.techknowlogick.com/xgo"
)
