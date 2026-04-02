package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestRequestConnectorFields(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": `olap_connector: duckdb`,
		},
	})
	s := newSession(t, rt, instanceID)

	t.Run("success", func(t *testing.T) {
		var res *ai.RequestConnectorFieldsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RequestConnectorFieldsName, &res, &ai.RequestConnectorFieldsArgs{
			Driver: "clickhouse",
			MissingFields: []string{
				"username",
				" password ",
				"username",
			},
			Message:       "Need credentials for localhost:9000",
			RelatedErrors: []string{" auth failed ", "auth failed"},
			ResourceName:  "clickhouse_local",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "connector_fields", res.HandoffKind)
		require.Equal(t, "clickhouse", res.Driver)
		require.Equal(t, []string{"username", "password"}, res.MissingFields)
		require.Equal(t, "Need credentials for localhost:9000", res.Message)
		require.Equal(t, "clickhouse_local", res.ResourceName)
		require.Equal(t, []string{"auth failed"}, res.RelatedErrors)
	})

	t.Run("missing driver", func(t *testing.T) {
		var res *ai.RequestConnectorFieldsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RequestConnectorFieldsName, &res, &ai.RequestConnectorFieldsArgs{
			MissingFields: []string{"password"},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "driver is required")
	})

	t.Run("missing fields", func(t *testing.T) {
		var res *ai.RequestConnectorFieldsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RequestConnectorFieldsName, &res, &ai.RequestConnectorFieldsArgs{
			Driver:        "s3",
			MissingFields: nil,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing_fields")
	})

	t.Run("invalid field key", func(t *testing.T) {
		var res *ai.RequestConnectorFieldsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.RequestConnectorFieldsName, &res, &ai.RequestConnectorFieldsArgs{
			Driver:        "s3",
			MissingFields: []string{"AWS_KEY"},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid missing_fields")
	})
}
