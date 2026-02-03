package gemini

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestDriverRegistered(t *testing.T) {
	d, ok := drivers.Drivers["gemini"]
	require.True(t, ok, "gemini driver should be registered")
	require.NotNil(t, d)

	spec := d.Spec()
	require.Equal(t, "Gemini", spec.DisplayName)
	require.True(t, spec.ImplementsAI)
}

func TestConfigDefaults(t *testing.T) {
	c := &configProperties{}
	require.Equal(t, "gemini-3-pro", c.getModel())

	c.Model = "custom-model"
	require.Equal(t, "custom-model", c.getModel())
}
