# Runtime pgwire adapter

This package exposes runtime Metrics SQL through the PostgreSQL wire protocol.
It maps runtime schemas and values to PostgreSQL OIDs and encodings, and builds
an in-memory DuckDB catalog containing the metrics views visible to the current
security claims. The catalog compatibility rewrites are intentionally scoped to
PostgreSQL introspection performed by psql, Superset, and Metabase.

Transport, authentication framing, prepared statements, portals, and
cancellation are implemented in `runtime/pkg/pgwire`.
