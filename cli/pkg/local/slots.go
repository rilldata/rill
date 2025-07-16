package local

import "github.com/rilldata/rill/cli/pkg/cmdutil"

// DefaultProdSlots returns the default number of slots for production environments.
//
// A slot represents the following resources:
//   - 1 CPU core
//   - 4 GB of memory
//   - 40 GB of storage
func DefaultProdSlots(ch *cmdutil.Helper) int {
	if ch.IsDev() {
		return 1
	}
	return 4
}
