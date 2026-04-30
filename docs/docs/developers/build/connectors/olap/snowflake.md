---
title: Snowflake
description: Power Rill dashboards using Snowflake
sidebar_label: Snowflake
sidebar_position: 24
---

[Snowflake](https://docs.snowflake.com) is a cloud data platform known for its scalability, ease of use, and separation of storage and compute. Rill supports connecting to Snowflake as a live connector, allowing you to build metrics views and dashboards directly on top of existing Snowflake tables — no data movement required.

:::tip Snowflake as a Live Connector vs. Data Source Connector
Rill supports Snowflake in two distinct modes:

- **OLAP (Live Connector)** — Rill queries Snowflake directly at dashboard load time. No data is ingested into Rill. Use this when your data is already modeled and optimized in Snowflake and you want Rill as a visual layer on top. Set `olap_connector: snowflake` in `rill.yaml`.

- **Data Source Connector** — Rill extracts data from Snowflake and ingests it into its embedded engine (DuckDB or ClickHouse). Use this when you want Rill to manage the data pipeline, apply transformations via SQL models, or combine Snowflake data with other sources. See the [Snowflake data source docs](/developers/build/connectors/data-source/snowflake).

In general, use the live connector if your Snowflake tables are already production-ready and large. Use data source ingestion if you need to transform, join, or enrich the data before building dashboards.
:::

## Connect to Snowflake

After selecting "Add Data", select Snowflake and fill in your connection parameters. This will automatically create the `snowflake.yaml` file in your `connectors` directory and populate the `.env` file with your credentials.

For more information on supported parameters, see our [Snowflake connector YAML reference docs](/reference/project-files/connectors#snowflake).

```yaml
type: connector
driver: snowflake

dsn: "{{ .env.SNOWFLAKE_DSN }}"
```

:::tip Finding your account identifier
Your account identifier appears in your Snowflake URL — it's everything before `.snowflakecomputing.com`. For example, if your URL is `https://xy12345.us-east-1.snowflakecomputing.com`, your account identifier is `xy12345.us-east-1`.
:::

### Key-Pair Authentication (Recommended)

For production use, Snowflake recommends key-pair authentication over passwords. Generate an unencrypted PKCS#8 private key and base64-encode it:

```bash
# Generate a 2048-bit unencrypted PKCS#8 private key
openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt -out rsa_key.p8

# Base64 URL-encode for use in Rill
base64 -w 0 rsa_key.p8
```

Then set the encoded key in your connector:

```yaml
type: connector
driver: snowflake

account: "<ACCOUNT_IDENTIFIER>"
user: "<USERNAME>"
private_key: "{{ .env.SNOWFLAKE_PRIVATE_KEY }}"
```

See [Snowflake's key-pair authentication docs](https://docs.snowflake.com/en/guide/key-pair-auth) for full setup instructions.

## Build a Metrics View on Snowflake

Once connected, set Snowflake as the OLAP connector in your `rill.yaml`:

```yaml
olap_connector: snowflake
```

Then create a metrics view that references a Snowflake table:

```yaml
type: metrics_view

connector: snowflake
database: MY_DATABASE
database_schema: MY_SCHEMA
model: MY_TABLE

timeseries: created_at
dimensions:
  - column: region
  - column: category
measures:
  - name: total_revenue
    expression: SUM(revenue)
```

:::note
Rill supports metrics views directly on Snowflake as a live connector. Incremental models and partitioned ingestion are not supported in live connector mode.
:::

### Caching Query Results

By default, dashboard queries against a Snowflake metrics view run live against your warehouse on every interaction. For dashboards with many concurrent users or repeated drill-downs, this can drive up warehouse compute costs and slow down the experience. Enable caching to reuse query results between users until the underlying data changes.

Caching is configured under the `cache` block on the metrics view. Because Snowflake is an external/live connector, caching is **off by default** — opt in by setting `cache.key_sql` (or `cache.enabled: true`).

```yaml
type: metrics_view

connector: snowflake
database: MY_DATABASE
database_schema: MY_SCHEMA
model: MY_TABLE

timeseries: created_at
dimensions:
  - column: region
measures:
  - name: total_revenue
    expression: SUM(revenue)

cache:
  key_sql: SELECT MAX(created_at) FROM MY_DATABASE.MY_SCHEMA.MY_TABLE
  key_ttl: 5m
```

Rill periodically runs `key_sql` against Snowflake (re-evaluated at most once per `key_ttl`) and uses the returned scalar value as the cache key. When the value changes — for example because a new row landed — the cache is invalidated and the next query repopulates it.

**Pros**

- **Lower Snowflake spend.** Repeat queries (multiple users on the same dashboard, back-and-forth filtering) are served from Rill's cache instead of hitting the warehouse.
- **Faster dashboards.** Cached results return in milliseconds; no warehouse warm-up or queue time.
- **Safe invalidation.** `key_sql` lets you tie cache freshness directly to your data — typically the max event timestamp or an ingest version column — so users never see stale results past your `key_ttl` window.

**Cons**

- **Up to `key_ttl` of staleness.** Between `key_sql` evaluations, new data in Snowflake will not appear on the dashboard. Pick a TTL that matches your data freshness SLA.
- **`key_sql` itself runs on Snowflake.** Make sure it's a cheap query (e.g. a `MAX()` on a clustered column). A slow `key_sql` will defeat the latency benefit.
- **Memory usage.** Cached results live in Rill's result cache; very high-cardinality dashboards with many distinct queries will evict older entries.
- **Not suitable for sub-minute freshness.** If you need near-real-time data, leave caching disabled and rely on Snowflake's own result cache instead.
