package gemini

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemperatureOnlyWhenConfigured(t *testing.T) {
	// Google recommends the default temperature for Gemini 3 models, and lower values can cause looping.
	require.Nil(t, (&configProperties{}).getTemperature())

	temperature := 0.5
	require.InDelta(t, 0.5, *(&configProperties{Temperature: &temperature}).getTemperature(), 1e-6)
}
