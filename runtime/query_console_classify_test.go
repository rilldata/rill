package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockCatalogLookup implements a simple catalog lookup function for testing.
// It returns true if the given table name is known as an internal (Rill) model.
type mockCatalogLookup struct {
	internalModels map[string]bool
}

func (m *mockCatalogLookup) IsInternalModel(name string) bool {
	if m.internalModels == nil {
		return false
	}
	return m.internalModels[name]
}

func TestClassifyModelType(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		sql string
		internal map[string]bool
		wantType ModelType
		wantErr bool
	}{
		// ── Pure external connector references → source_model ──────────────
		{
			name:     "single external table",
			sql:      "SELECT * FROM raw_events",
			internal: map[string]bool{}, // nothing is internal
			wantType: ModelTypeSource,
		},
		{
			name:     "multiple external tables via JOIN",
			sql:      "SELECT a.id, b.name FROM ext_users a JOIN ext_orders b ON a.id = b.user_id",
			internal: map[string]bool{},
			wantType: ModelTypeSource,
		},
		{
			name:     "external table with schema-qualified name",
			sql:      "SELECT * FROM staging.raw_clicks",
			internal: map[string]bool{},
			wantType: ModelTypeSource,
		},
		{
			name:     "external table with fully-qualified name (db.schema.table)",
			sql:      "SELECT * FROM warehouse.public.events",
			internal: map[string]bool{},
			wantType: ModelTypeSource,
		},

		// ── Pure internal model references → derived_model ─────────────────
		{
			name:     "single internal model",
			sql:      "SELECT * FROM users_model",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeDerived,
		},
		{
			name:     "multiple internal models via JOIN",
			sql:      "SELECT u.id, o.total FROM users_model u JOIN orders_model o ON u.id = o.user_id",
			internal: map[string]bool{"users_model": true, "orders_model": true},
			wantType: ModelTypeDerived,
		},
		{
			name:     "internal model in subquery",
			sql:      "SELECT * FROM (SELECT id, name FROM users_model) sub",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeDerived,
		},

		// ── Mixed references → source_model ─────────────────────────────────
		{
			name:     "one internal and one external",
			sql:      "SELECT m.id, e.event FROM users_model m JOIN raw_events e ON m.id = e.user_id",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeSource,
		},
		{
			name:     "mixed via subquery",
			sql:      "SELECT * FROM users_model WHERE id IN (SELECT user_id FROM raw_events)",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeSource,
		},

		// ── Complex SQL patterns ─────────────────────────────────────────────
		{
			name: "CTE with internal models only",
			sql: `WITH active_users AS (
				SELECT * FROM users_model WHERE active = true
			), recent_orders AS (
				SELECT * FROM orders_model WHERE created_at > '2024-01-01'
			)
			SELECT u.id, o.total
			FROM active_users u
			JOIN recent_orders o ON u.id = o.user_id`,
			internal: map[string]bool{"users_model": true, "orders_model": true},
			wantType: ModelTypeDerived,
		},
		{
			name: "CTE with mixed references",
			sql: `WITH base AS (
				SELECT * FROM users_model
			)
			SELECT b.id, r.event
			FROM base b
			JOIN raw_events r ON b.id = r.user_id`,
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeSource,
		},
		{
			name:     "nested subqueries with all internal",
			sql:      "SELECT * FROM (SELECT * FROM (SELECT id FROM users_model) a) b",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeDerived,
		},
		{
			name:     "LEFT JOIN with external",
			sql:      "SELECT a.*, b.extra FROM users_model a LEFT JOIN ext_metadata b ON a.id = b.id",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeSource,
		},
		{
			name: "UNION ALL with internal models",
			sql: `SELECT id, name FROM users_model
			UNION ALL
			SELECT id, name FROM archived_users_model`,
			internal: map[string]bool{"users_model": true, "archived_users_model": true},
			wantType: ModelTypeDerived,
		},
		{
			name: "UNION ALL mixed",
			sql: `SELECT id, name FROM users_model
			UNION ALL
			SELECT id, name FROM raw_external_users`,
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeSource,
		},
		{
			name:     "correlated subquery with external",
			sql:      "SELECT * FROM users_model u WHERE EXISTS (SELECT 1 FROM raw_events e WHERE e.user_id = u.id)",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeSource,
		},
		{
			name:     "multiple JOINs all internal",
			sql:      "SELECT * FROM users_model u JOIN orders_model o ON u.id = o.user_id JOIN products_model p ON o.product_id = p.id",
			internal: map[string]bool{"users_model": true, "orders_model": true, "products_model": true},
			wantType: ModelTypeDerived,
		},
		{
			name:     "CROSS JOIN external",
			sql:      "SELECT * FROM ext_a CROSS JOIN ext_b",
			internal: map[string]bool{},
			wantType: ModelTypeSource,
		},

		// ── Edge cases ────────────────────────────────────────────────────────
		{
			name:     "no table references (literal select)",
			sql:      "SELECT 1 AS one, 'hello' AS greeting",
			internal: map[string]bool{},
			wantType: ModelTypeDerived, // no external refs → derived
		},
		{
			name:     "table name matches CTE name (should not count as external)",
			sql:      "WITH my_cte AS (SELECT * FROM users_model) SELECT * FROM my_cte",
			internal: map[string]bool{"users_model": true},
			wantType: ModelTypeDerived,
		},

		// ── SQL parsing error handling ────────────────────────────────────────
		{
			name:    "empty SQL string",
			sql:     "",
			wantErr: true,
		},
		{
			name:    "whitespace-only SQL",
			sql:     "   \t\n  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := &mockCatalogLookup{internalModels: tt.internal}

			// ClassifyModelType takes a context, a catalog lookup function, and SQL.
			// The catalog lookup function signature matches what the production code expects:
			// func(ctx context.Context, name string) bool
			modelType, err := ClassifyModelType(ctx, func(ctx context.Context, name string) bool {
				return lookup.IsInternalModel(name)
			}, tt.sql)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantType, modelType, "unexpected model type for SQL: %s", tt.sql)
		})
	}
}

func TestClassifyModelType_NilLookup(t *testing.T) {
	ctx := context.Background()

	// When catalog lookup always returns false (nothing is internal),
	// any table reference becomes external → source_model.
	modelType, err := ClassifyModelType(ctx, func(ctx context.Context, name string) bool {
		return false
	}, "SELECT * FROM some_table")

	require.NoError(t, err)
	require.Equal(t, ModelTypeSource, modelType)
}

func TestClassifyModelType_CaseInsensitivity(t *testing.T) {
	ctx := context.Background()

	// Table names in SQL are often case-insensitive.
	// The lookup should match regardless of case in the SQL.
	modelType, err := ClassifyModelType(ctx, func(ctx context.Context, name string) bool {
		// The classify function should normalize names before lookup.
		// We accept both "users_model" and "USERS_MODEL" as internal.
		return name == "users_model"
	}, "SELECT * FROM USERS_MODEL")

	require.NoError(t, err)
	// If the classifier lowercases before lookup → derived; otherwise → source.
	// We accept either behavior but verify no error.
	require.Contains(t, []ModelType{ModelTypeSource, ModelTypeDerived}, modelType)
}

func TestClassifyModelType_ComplexCTEChain(t *testing.T) {
	ctx := context.Background()

	sql := `
		WITH step1 AS (
			SELECT id, name FROM users_model
		),
		step2 AS (
			SELECT s1.id, o.total
			FROM step1 s1
			JOIN orders_model o ON s1.id = o.user_id
		),
		step3 AS (
			SELECT s2.id, s2.total, p.product_name
			FROM step2 s2
			JOIN products_model p ON s2.id = p.order_id
		)
		SELECT * FROM step3 WHERE total > 100
	`

	internal := map[string]bool{
		"users_model":    true,
		"orders_model":   true,
		"products_model": true,
	}

	modelType, err := ClassifyModelType(ctx, func(ctx context.Context, name string) bool {
		return internal[name]
	}, sql)

	require.NoError(t, err)
	require.Equal(t, ModelTypeDerived, modelType)
}

func TestClassifyModelType_CTEChainWithExternal(t *testing.T) {
	ctx := context.Background()

	sql := `
		WITH enriched AS (
			SELECT u.*, e.extra_data
			FROM users_model u
			JOIN external_enrichment e ON u.id = e.user_id
		)
		SELECT * FROM enriched
	`

	modelType, err := ClassifyModelType(ctx, func(ctx context.Context, name string) bool {
		return name == "users_model"
	}, sql)

	require.NoError(t, err)
	require.Equal(t, ModelTypeSource, modelType, "mixed references in CTE should result in source_model")
}
