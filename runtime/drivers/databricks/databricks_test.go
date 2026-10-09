package databricks

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

// TestResolveDSN pins that resolveDSN never selects a backend itself: the DSN carries no
// useKernel param (the driver switches Lakehouse//RT to SEA on its own), and a raw DSN is
// passed through untouched, so useKernel=true in a raw DSN still forces SEA.
func TestResolveDSN(t *testing.T) {
	t.Run("params: no useKernel", func(t *testing.T) {
		c := &configProperties{Host: "h.cloud.databricks.com", HTTPPath: "/sql/1.0/warehouses/w", Token: "t"}
		require.NotContains(t, c.resolveDSN(), "useKernel")
	})

	t.Run("dsn: passed through unchanged", func(t *testing.T) {
		dsn := "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w?useKernel=true"
		c := &configProperties{DSN: dsn}
		require.Equal(t, dsn, c.resolveDSN())
	})
}

// TestRemovedUseKernelProperty checks that connectors still setting the removed
// use_kernel property keep decoding.
func TestRemovedUseKernelProperty(t *testing.T) {
	conf := &configProperties{}
	err := mapstructure.WeakDecode(map[string]any{"host": "h", "http_path": "/p", "token": "t", "use_kernel": true}, conf)
	require.NoError(t, err)
	require.NoError(t, conf.validate())
	require.NotContains(t, conf.resolveDSN(), "useKernel")
}
