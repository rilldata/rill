# Runtime pgwire adapter

This package exposes runtime Metrics SQL through the PostgreSQL wire protocol.
It maps runtime schemas and values to PostgreSQL OIDs and encodings, and builds
an in-memory DuckDB catalog containing the metrics views visible to the current
security claims. The catalog compatibility rewrites are intentionally scoped to
PostgreSQL introspection performed by psql, Superset, and Metabase.

Transport, authentication framing, prepared statements, portals, and
cancellation are implemented in `runtime/pkg/pgwire`.

Statement description uses the Metrics SQL compiler and the OLAP schema API,
without fetching query result rows. Catalog columns use the validated metrics
view dimension and measure types. Catalog SQL runs in a separate in-memory
DuckDB with external access disabled and configuration locked; only a single
SELECT (or an explicitly supported SHOW probe) is accepted. The catalog
database is built once per session and rebuilt when a metrics view changes.
Each catalog query runs on its own connection, which owns the connection-local
default schema and `pg_matviews` temporary table.

Routing, parameter inference, and binding share a parsed SQL representation.
Catalog routing recognizes relation references, excluding column names and
quoted values. Catalog parameters use native DuckDB binding with declared types;
Metrics SQL parameters use MySQL-compatible literals. LIMIT and OFFSET parameters
are inferred as integers and use zero placeholders during schema discovery.

Catalog compatibility rewrites match complete SQL tokens, preserving quoted text
and comments. Function probes preserve caller aliases and only add their default
column names for standalone, unaliased SELECT projections.
