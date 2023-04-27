package admin

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	x := "github"
	y := "github"

	require.True(t, reflect.DeepEqual(&x, &y))
}
