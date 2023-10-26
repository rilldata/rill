package queries

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTickIncrementNaN_AnyNaN(t *testing.T) {
	require.True(t, math.IsNaN(tickIncrement(math.NaN(), 1, 1)))
	require.True(t, math.IsNaN(tickIncrement(0, math.NaN(), 1)))
	require.True(t, math.IsNaN(tickIncrement(0, 1, math.NaN())))
	require.True(t, math.IsNaN(tickIncrement(math.NaN(), math.NaN(), 1)))
	require.True(t, math.IsNaN(tickIncrement(0, math.NaN(), math.NaN())))
	require.True(t, math.IsNaN(tickIncrement(math.NaN(), 1, math.NaN())))
	require.True(t, math.IsNaN(tickIncrement(math.NaN(), math.NaN(), math.NaN())))
}

func TestTickIncrementNaN_StartEqualsStop(t *testing.T) {
	require.True(t, math.IsNaN(tickIncrement(1, 1, -1)))
	require.True(t, math.IsNaN(tickIncrement(1, 1, 0)))
	require.True(t, math.IsNaN(tickIncrement(1, 1, math.NaN())))
	require.Equal(t, -math.Inf(1), tickIncrement(1, 1, 1))
	require.Equal(t, -math.Inf(1), tickIncrement(1, 1, 10))
}

func TestTickIncrementZeroOrInf_CountNotPositive(t *testing.T) {
	require.Equal(t, math.Inf(1), tickIncrement(0, 1, -1))
	require.Equal(t, math.Inf(1), tickIncrement(0, 1, 0))
}

func TestTickIncrementInf_CountInf(t *testing.T) {
	require.Equal(t, -math.Inf(1), tickIncrement(0, 1, math.Inf(1)))
}

func TestTickIncrementCountPlus1_StartLessThanStop(t *testing.T) {
	require.Equal(t, -10.0, tickIncrement(0, 1, 10))
	require.Equal(t, -10.0, tickIncrement(0, 1, 9))
	require.Equal(t, -10.0, tickIncrement(0, 1, 8))
	require.Equal(t, -5.0, tickIncrement(0, 1, 7))
	require.Equal(t, -5.0, tickIncrement(0, 1, 6))
	require.Equal(t, -5.0, tickIncrement(0, 1, 5))
	require.Equal(t, -5.0, tickIncrement(0, 1, 4))
	require.Equal(t, -2.0, tickIncrement(0, 1, 3))
	require.Equal(t, -2.0, tickIncrement(0, 1, 2))
	require.Equal(t, 1.0, tickIncrement(0, 1, 1))
	require.Equal(t, 1.0, tickIncrement(0, 10, 10))
	require.Equal(t, 1.0, tickIncrement(0, 10, 9))
	require.Equal(t, 1.0, tickIncrement(0, 10, 8))
	require.Equal(t, 2.0, tickIncrement(0, 10, 7))
	require.Equal(t, 2.0, tickIncrement(0, 10, 6))
	require.Equal(t, 2.0, tickIncrement(0, 10, 5))
	require.Equal(t, 2.0, tickIncrement(0, 10, 4))
	require.Equal(t, 5.0, tickIncrement(0, 10, 3))
	require.Equal(t, 5.0, tickIncrement(0, 10, 2))
	require.Equal(t, 10.0, tickIncrement(0, 10, 1))
	require.Equal(t, 2.0, tickIncrement(-10, 10, 10))
	require.Equal(t, 2.0, tickIncrement(-10, 10, 9))
	require.Equal(t, 2.0, tickIncrement(-10, 10, 8))
	require.Equal(t, 2.0, tickIncrement(-10, 10, 7))
	require.Equal(t, 5.0, tickIncrement(-10, 10, 6))
	require.Equal(t, 5.0, tickIncrement(-10, 10, 5))
	require.Equal(t, 5.0, tickIncrement(-10, 10, 4))
	require.Equal(t, 5.0, tickIncrement(-10, 10, 3))
	require.Equal(t, 10.0, tickIncrement(-10, 10, 2))
	require.Equal(t, 20.0, tickIncrement(-10, 10, 1))
}
