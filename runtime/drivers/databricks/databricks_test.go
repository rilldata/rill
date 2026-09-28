package databricks

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResolveDSN pins the backend-selection behavior of resolveDSN:
//   - by default (use_kernel unset) the DSN carries no useKernel param, so ordinary
//     DBSQL warehouses keep using the Thrift backend (backward compatible);
//   - with use_kernel=true the DSN carries useKernel=true, selecting the SEA/kernel
//     backend required by Lakehouse//RT;
//   - a raw DSN is passed through untouched unless use_kernel opts in, and is never
//     given a duplicate useKernel param.
func TestResolveDSN(t *testing.T) {
	t.Run("params: default has no useKernel", func(t *testing.T) {
		c := &configProperties{Host: "h.cloud.databricks.com", HTTPPath: "/sql/1.0/warehouses/w", Token: "t"}
		require.NotContains(t, c.resolveDSN(), "useKernel")
	})

	t.Run("params: use_kernel adds useKernel=true", func(t *testing.T) {
		c := &configProperties{Host: "h.cloud.databricks.com", HTTPPath: "/sql/1.0/warehouses/w", Token: "t", UseKernel: true}
		require.Contains(t, c.resolveDSN(), "useKernel=true")
	})

	t.Run("dsn: passed through unchanged by default", func(t *testing.T) {
		dsn := "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w"
		c := &configProperties{DSN: dsn}
		require.Equal(t, dsn, c.resolveDSN())
	})

	t.Run("dsn: use_kernel appends with ? when no query", func(t *testing.T) {
		c := &configProperties{DSN: "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w", UseKernel: true}
		require.Equal(t, "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w?useKernel=true", c.resolveDSN())
	})

	t.Run("dsn: use_kernel appends with & when query exists", func(t *testing.T) {
		c := &configProperties{DSN: "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w?catalog=main", UseKernel: true}
		require.Equal(t, "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w?catalog=main&useKernel=true", c.resolveDSN())
	})

	t.Run("dsn: no duplicate useKernel", func(t *testing.T) {
		c := &configProperties{DSN: "token:tok@h.cloud.databricks.com:443/sql/1.0/warehouses/w?useKernel=true", UseKernel: true}
		require.Equal(t, 1, strings.Count(c.resolveDSN(), "useKernel="))
	})
}

// TestRTRequiresSEA pins the Lakehouse//RT autodetect trigger: only the server's
// "not supported for Thrift protocol" message (however wrapped) flips to SEA; other
// connection errors must not.
func TestRTRequiresSEA(t *testing.T) {
	require.True(t, rtRequiresSEA(errors.New("Lakehouse/RT is not supported for Thrift protocol. Please update...")))
	require.True(t, rtRequiresSEA(fmt.Errorf("wrap: %w", errors.New("...is not supported for Thrift protocol"))))
	require.False(t, rtRequiresSEA(errors.New("HTTP 403: Invalid access token")))
	require.False(t, rtRequiresSEA(errors.New("connection refused")))
	require.False(t, rtRequiresSEA(nil))
}

func TestWithUseKernel(t *testing.T) {
	require.Equal(t, "dsn?useKernel=true", withUseKernel("dsn"))
	require.Equal(t, "dsn?x=1&useKernel=true", withUseKernel("dsn?x=1"))
	require.Equal(t, "dsn?useKernel=true", withUseKernel("dsn?useKernel=true"))                  // no dup
	require.Equal(t, "dsn?useKernel=true", withUseKernel("dsn?useKernel=false"))                 // overrides false
	require.Equal(t, "dsn?a=1&b=2&useKernel=true", withUseKernel("dsn?a=1&useKernel=false&b=2")) // drops mid, re-adds
}
