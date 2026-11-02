package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestSQLLimit(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"foo.sql": "SELECT range AS val FROM range(100)",
		},
		Variables: map[string]string{
			"rill.interactive_sql_row_limit": "10",
		},
	})

	cases := []struct {
		name      string
		sql       string
		limit     int
		wantRows  int
		wantError string
	}{
		{
			name:      "bad unlimited",
			sql:       "SELECT * FROM foo",
			wantError: "result cap exceeded",
		},
		{
			name:      "bad explicit limit",
			sql:       "SELECT * FROM foo",
			limit:     15,
			wantError: "exceeds the maximum interactive limit",
		},
		{
			name:     "good unlimited",
			sql:      "SELECT 1",
			wantRows: 1,
		},
		{
			name:     "good normal limit",
			sql:      "SELECT * FROM foo LIMIT 5",
			wantRows: 5,
		},
		{
			name:     "good explicit limit",
			sql:      "SELECT * FROM foo",
			limit:    5,
			wantRows: 5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
				InstanceID:         instanceID,
				Resolver:           "sql",
				ResolverProperties: map[string]any{"sql": tc.sql, "limit": tc.limit},
				Claims:             &runtime.SecurityClaims{SkipChecks: true},
			})
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			defer res.Close()

			var rows []map[string]any
			for {
				row, err := res.Next()
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
				rows = append(rows, row)
			}

			if tc.limit > 0 {
				require.Equal(t, tc.wantRows, len(rows))
			}
		})
	}
}

func TestSimpleSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "simple_sql_api", &runtime.SecurityClaims{})
	require.NoError(t, err)

	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               nil,
		Claims:             &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 5, len(rows[0]))
	require.Equal(t, 4.09, rows[0]["bid_price"])
	require.Equal(t, "msn.com", rows[0]["domain"])
	require.Equal(t, float64(4000), rows[0]["id"])
	require.Equal(t, nil, rows[0]["publisher"])
	require.Equal(t, "2022-03-05T14:49:50.459Z", rows[0]["timestamp"])
}

func TestTemplateSQLApi(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "templated_sql_api", &runtime.SecurityClaims{})
	require.NoError(t, err)

	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               map[string]any{"domain": "sports.yahoo.com"},
		Claims:             &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 5, len(rows[0]))
	require.Equal(t, 1.81, rows[0]["bid_price"])
	require.Equal(t, "sports.yahoo.com", rows[0]["domain"])
	require.Equal(t, float64(9000), rows[0]["id"])
	require.Equal(t, "Yahoo", rows[0]["publisher"])
	require.Equal(t, "2022-02-09T11:58:12.475Z", rows[0]["timestamp"])
	for _, row := range rows {
		require.Equal(t, "sports.yahoo.com", row["domain"])
	}
}

func TestTemplateSQLApi2(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	api, err := rt.APIForName(context.Background(), instanceID, "templated_sql_api_2", &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "msn.com"}})
	require.NoError(t, err)

	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           api.Spec.Resolver,
		ResolverProperties: api.Spec.ResolverProperties.AsMap(),
		Args:               map[string]any{"pageSize": 5},
		Claims:             &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "msn.com"}},
	})
	require.NoError(t, err)
	defer res.Close()

	require.NotNil(t, res)

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))
	require.Equal(t, 5, len(rows))
	require.Equal(t, 5, len(rows[0]))
	require.Equal(t, 4.09, rows[0]["bid_price"])
	require.Equal(t, "msn.com", rows[0]["domain"])
	require.Equal(t, float64(4000), rows[0]["id"])
	require.Equal(t, nil, rows[0]["publisher"])
	require.Equal(t, "2022-03-05T14:49:50.459Z", rows[0]["timestamp"])
	for _, row := range rows {
		require.Equal(t, "msn.com", row["domain"])
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// TestSQLUntrustedPropertiesHideVariables verifies that templates in resolver properties supplied by an API caller
// cannot read project variables unless the caller may read sensitive instance info,
// while templates declared in the project always can.
// TestSQLTemplateVariables verifies that only templates declared in the project's source files may access project variables,
// whoever the caller is.
func TestSQLTemplateVariables(t *testing.T) {
	const secret = "s3cret-value"
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:     map[string]string{"rill.yaml": ""},
		Variables: map[string]string{"secret": secret},
	})

	cases := []struct {
		name            string
		sql             string
		projectDeclared bool
		claims          *runtime.SecurityClaims
		want            string
		wantErr         string
	}{
		{
			name:            "project-declared properties see variables",
			sql:             "SELECT '{{ .env.secret }}' AS v",
			projectDeclared: true,
			claims:          &runtime.SecurityClaims{},
			want:            secret,
		},
		{
			name:    "caller-supplied properties cannot reference variables",
			sql:     "SELECT '{{ .env.secret }}' AS v",
			claims:  &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics, runtime.ReadResolvers}},
			wantErr: "project variables are only available to templates declared in the project files",
		},
		{
			name:    "caller-supplied properties cannot reference variables even with ReadInstance",
			sql:     "SELECT '{{ .env.secret }}' AS v",
			claims:  &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ReadInstance}},
			wantErr: "project variables are only available to templates declared in the project files",
		},
		{
			name:    "caller-supplied properties cannot reference variables even when checks are skipped",
			sql:     "SELECT '{{ .env.secret }}' AS v",
			claims:  &runtime.SecurityClaims{SkipChecks: true},
			wantErr: "project variables are only available to templates declared in the project files",
		},
		{
			name:   "caller-supplied properties without variables resolve",
			sql:    "SELECT '{{ .args.x }}' AS v",
			claims: &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ReadAPI, runtime.ReadMetrics, runtime.ReadResolvers}},
			want:   "<no value>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
				InstanceID:         instanceID,
				Resolver:           "sql",
				ResolverProperties: map[string]any{"sql": tc.sql},
				ProjectDeclared:    tc.projectDeclared,
				Claims:             tc.claims,
			})
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.NotContains(t, err.Error(), secret)
				return
			}
			require.NoError(t, err)
			defer res.Close()

			row, err := res.Next()
			require.NoError(t, err)
			require.Equal(t, tc.want, row["v"])
		})
	}
}
