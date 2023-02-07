package container

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/stretchr/testify/require"
)

func TestUnboundedContainer(t *testing.T) {
	c, err := NewUnbounded[int]()
	require.NoError(t, err)

	i := 0
	for ; i < 50 && !c.Full(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 50, i)
	require.Equal(t, arrayutil.RangeInt(0, 50, false), c.Items())
}
