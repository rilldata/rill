package clickhouse

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNamedCollectionName(t *testing.T) {
	require.Equal(t, "rill_my_bucket", NamedCollectionName("my_bucket"))
}

func TestIsSupportedNamedCollectionDriver(t *testing.T) {
	for _, d := range []string{"s3", "gcs", "azure", "mysql", "postgres"} {
		require.True(t, IsSupportedNamedCollectionDriver(d), d)
	}
	for _, d := range []string{"duckdb", "clickhouse", "bigquery", "https", ""} {
		require.False(t, IsSupportedNamedCollectionDriver(d), d)
	}
}

func TestBuildNamedCollectionParams_S3(t *testing.T) {
	params, err := BuildNamedCollectionParams("s3", map[string]any{
		"aws_access_key_id":     "AKIA",
		"aws_secret_access_key": "secret",
		"region":                "us-east-1",
	})
	require.NoError(t, err)
	got := paramsAsMap(params)
	require.Equal(t, map[string]string{
		"access_key_id":     "AKIA",
		"secret_access_key": "secret",
		"region":            "us-east-1",
	}, got)
}

func TestBuildNamedCollectionParams_GCS_RequiresHMAC(t *testing.T) {
	_, err := BuildNamedCollectionParams("gcs", map[string]any{
		"google_application_credentials": "{...}",
	})
	require.ErrorIs(t, err, ErrGCSRequiresHMAC)

	params, err := BuildNamedCollectionParams("gcs", map[string]any{
		"key_id": "GOOG-KEY",
		"secret": "secret",
	})
	require.NoError(t, err)
	got := paramsAsMap(params)
	require.Equal(t, "GOOG-KEY", got["access_key_id"])
	require.Equal(t, "secret", got["secret_access_key"])
	require.Equal(t, "https://storage.googleapis.com", got["endpoint"])
}

func TestBuildNamedCollectionParams_Azure(t *testing.T) {
	params, err := BuildNamedCollectionParams("azure", map[string]any{
		"azure_storage_account": "myacct",
		"azure_storage_key":     "mykey",
	})
	require.NoError(t, err)
	got := paramsAsMap(params)
	require.Equal(t, "myacct", got["account_name"])
	require.Equal(t, "mykey", got["account_key"])
}

func TestBuildNamedCollectionParams_Postgres(t *testing.T) {
	params, err := BuildNamedCollectionParams("postgres", map[string]any{
		"host":     "db.internal",
		"port":     "5433",
		"user":     "rill",
		"password": "p@ss",
		"dbname":   "analytics",
	})
	require.NoError(t, err)
	got := paramsAsMap(params)
	require.Equal(t, "db.internal:5433", got["host"])
	require.Equal(t, "rill", got["user"])
	require.Equal(t, "p@ss", got["password"])
	require.Equal(t, "analytics", got["database"])
}

func TestBuildNamedCollectionParams_MySQL(t *testing.T) {
	params, err := BuildNamedCollectionParams("mysql", map[string]any{
		"host":     "mysql.internal",
		"user":     "rill",
		"database": "analytics",
	})
	require.NoError(t, err)
	got := paramsAsMap(params)
	require.Equal(t, "mysql.internal:3306", got["host"])
	require.Equal(t, "rill", got["user"])
	require.Equal(t, "analytics", got["database"])
	_, hasPwd := got["password"]
	require.False(t, hasPwd)
}

func TestBuildNamedCollectionParams_UnsupportedDriver(t *testing.T) {
	_, err := BuildNamedCollectionParams("duckdb", map[string]any{})
	require.True(t, errors.Is(err, ErrUnsupportedNamedCollectionDriver))
}

func TestBuildCreateNamedCollectionSQL(t *testing.T) {
	params := []namedCollectionParam{
		{Key: "access_key_id", Value: "AKIA"},
		{Key: "secret_access_key", Value: "shh"},
	}
	sql, err := buildCreateNamedCollectionSQL("my_bucket", params, "")
	require.NoError(t, err)
	require.Equal(t, `CREATE OR REPLACE NAMED COLLECTION "rill_my_bucket" AS access_key_id = 'AKIA', secret_access_key = 'shh'`, sql)

	clusterSQL, err := buildCreateNamedCollectionSQL("my_bucket", params, "my_cluster")
	require.NoError(t, err)
	require.Equal(t, `CREATE OR REPLACE NAMED COLLECTION "rill_my_bucket" ON CLUSTER "my_cluster" AS access_key_id = 'AKIA', secret_access_key = 'shh'`, clusterSQL)
}

func TestBuildCreateNamedCollectionSQL_EscapesValues(t *testing.T) {
	// Single-quote in value must be escaped to prevent SQL injection.
	params := []namedCollectionParam{{Key: "password", Value: "ev'il"}}
	sql, err := buildCreateNamedCollectionSQL("c", params, "")
	require.NoError(t, err)
	require.Contains(t, sql, "password = 'ev''il'")
}

func TestBuildDropNamedCollectionSQL(t *testing.T) {
	require.Equal(t, `DROP NAMED COLLECTION IF EXISTS "rill_my_bucket"`, buildDropNamedCollectionSQL("my_bucket", ""))
	require.Equal(t, `DROP NAMED COLLECTION IF EXISTS "rill_my_bucket" ON CLUSTER "my_cluster"`, buildDropNamedCollectionSQL("my_bucket", "my_cluster"))
}

func TestDetectNamedCollectionRefs(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want []string
	}{
		{"none", "SELECT * FROM foo", []string{}},
		{"s3 simple", "SELECT * FROM s3(rill_my_bucket, url='s3://x/y')", []string{"my_bucket"}},
		{"postgresql", "SELECT * FROM postgresql(rill_pg, table='users')", []string{"pg"}},
		{"azure", "SELECT * FROM azureBlobStorage(rill_az, container='c', blob_path='*')", []string{"az"}},
		{"mysql case-insensitive", "SELECT * FROM MYSQL(rill_db, table='t')", []string{"db"}},
		{"s3Cluster", "SELECT * FROM s3Cluster('cluster', rill_my_bucket, url='s3://x/y')", []string{"my_bucket"}},
		{"multiple", "SELECT a.x FROM s3(rill_a, url='') a JOIN postgresql(rill_b, table='t') b ON a.id=b.id", []string{"a", "b"}},
		{"dedup", "SELECT * FROM s3(rill_a) UNION ALL SELECT * FROM s3(rill_a, url='other')", []string{"a"}},
		{"plain rill_ identifier (not a function call) is ignored", "SELECT rill_foo FROM bar", []string{}},
		{"url table function", "SELECT * FROM url(rill_https, format='CSV')", []string{"https"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectNamedCollectionRefs(tc.sql)
			if len(tc.want) == 0 {
				require.Empty(t, got)
				return
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func paramsAsMap(p []namedCollectionParam) map[string]string {
	m := make(map[string]string, len(p))
	for _, x := range p {
		m[x.Key] = x.Value
	}
	return m
}
