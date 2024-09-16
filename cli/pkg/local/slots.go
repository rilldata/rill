package local

import "github.com/rilldata/rill/cli/pkg/cmdutil"

func DefaultProdSlots(ch *cmdutil.Helper) int {
	if ch.IsDev() {
		return 1
	}
	return 2
}
