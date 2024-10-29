package rilltime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {
	rt, err := Parse("-1M/h")
	require.NoError(t, err)
	fmt.Println(rt.String())

	rt, err = Parse("-1M/h : @-3M")
	require.NoError(t, err)
	fmt.Println(rt.String())

	rt, err = Parse("-1M/h, now")
	require.NoError(t, err)
	fmt.Println(rt.String())

	rt, err = Parse("-1M/h, now : @-3M")
	require.NoError(t, err)
	fmt.Println(rt.String())

	rt, err = Parse("-1M/h, now : |h| @-3M")
	require.NoError(t, err)
	fmt.Println(rt.String())
}
