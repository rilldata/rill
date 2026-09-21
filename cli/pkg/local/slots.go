package local

import "github.com/rilldata/rill/cli/pkg/cmdutil"

// DefaultProdSlots returns the prod slot count to request from the admin API.
//
// Returns 0 for released builds, which signals the admin server to apply its own default.
// Dev builds request two slots to match the default for new deployments.
//
// A slot represents the following resources:
//   - 1 CPU core
//   - 4 GB of memory
//   - 40 GB of storage
func DefaultProdSlots(ch *cmdutil.Helper) int {
	if ch.IsDev() {
		return 2
	}
	return 0
}

// DefaultDevSlots returns the dev slot count to request from the admin API.
//
// Returns 0 for released builds, which signals the admin server to apply its own default.
// Dev builds request two slots to match the default for new deployments.
//
// A slot represents the following resources:
//   - 1 CPU core
//   - 4 GB of memory
//   - 40 GB of storage
func DefaultDevSlots(ch *cmdutil.Helper) int {
	if ch.IsDev() {
		return 2
	}
	return 0
}
