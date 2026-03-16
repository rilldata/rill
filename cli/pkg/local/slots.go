package local

import "github.com/rilldata/rill/cli/pkg/cmdutil"

// DefaultProdSlots returns the default number of slots for production environments.
// All plans start with 1 slot; users can manually increase after deployment.
//
// A slot represents the following resources:
//   - 1 CPU core
//   - 4 GB of memory
//   - 40 GB of storage
func DefaultProdSlots(_ *cmdutil.Helper) int {
	return 1
}
