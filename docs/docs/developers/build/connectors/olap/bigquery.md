---
title: Google BigQuery
description: Power Rill dashboards using BigQuery
sidebar_label: Google BigQuery
sidebar_position: 13
---

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets. Rill supports connecting to BigQuery as a live connector, allowing you to build metrics views and dashboards directly on top of existing BigQuery tables — no data movement required.

:::tip BigQuery as a Live Connector vs. Data Source Connector
Rill supports BigQuery in two distinct modes:

- **OLAP (Live Connector)** — Rill queries BigQuery directly at dashboard load time. No data is ingested into Rill. Use this when your data is already modeled and optimized in BigQuery and you want Rill as a visual layer on top. Set `olap_connector: bigquery` in `rill.yaml`.

- **Data Source Connector** — Rill extracts data from BigQuery and ingests it into its embedded engine (DuckDB or ClickHouse). Use this when you want Rill to manage the data pipeline, apply transformations via SQL models, or combine BigQuery data with other sources. See the [BigQuery data source docs](/developers/build/connectors/data-source/bigquery).

In general, use the live connector if your BigQuery tables are already production-ready and large. Use data source ingestion if you need to transform, join, or enrich the data before building dashboards.
:::

## Connect to BigQuery

After selecting "Add Data", select BigQuery and fill in your connection parameters. This will automatically create the `bigquery.yaml` file in your `connectors` directory and populate the `.env` file with your credentials.

For more information on supported parameters, see our [BigQuery connector YAML reference docs](/reference/project-files/connectors#bigquery).

```yaml
type: connector
driver: bigquery

project_id: "my-gcp-project"
google_application_credentials: "{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}"
```

:::tip Creating a Service Account
For production use, authenticate with a Google Cloud service account JSON key. The service account needs the `roles/bigquery.dataViewer`, `roles/bigquery.readSessionUser`, and `roles/bigquery.jobUser` roles. See [the appendix in the BigQuery data source docs](/developers/build/connectors/data-source/bigquery#how-to-create-a-service-account-using-the-google-cloud-console) for a step-by-step guide.
:::

### Application Default Credentials (Local Development)

For local development, you can omit `google_application_credentials` and Rill will use your local Google Cloud CLI credentials:

```bash
gcloud auth application-default login
```

```yaml
type: connector
driver: bigquery

project_id: "my-gcp-project"
```

:::warning Not suitable for Rill Cloud
Application Default Credentials only work locally. When deploying to Rill Cloud, always provide a service account JSON via `google_application_credentials`.
:::

### Controlling Query Costs

BigQuery bills based on bytes scanned. On the on-demand pricing model, we recommend setting `max_bytes_billed` on the connector to cap the amount scanned per dashboard query. Queries that exceed the limit fail with an error instead of running.

```yaml
type: connector
driver: bigquery

project_id: "my-gcp-project"
google_application_credentials: "{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}"
max_bytes_billed: 100000000000 # 100 GB
```

`max_bytes_billed` only applies to live connector dashboard queries. It is not enforced when ingesting data via the BigQuery data source connector. The default value of `0` disables the limit.

## Build a Metrics View on BigQuery

Once connected, set BigQuery as the OLAP connector in your `rill.yaml`:

```yaml
olap_connector: bigquery
```

Then create a metrics view that references a BigQuery table:

```yaml
type: metrics_view

connector: bigquery
database: my-gcp-project    # BigQuery project (optional; defaults to the connector's project_id)
database_schema: my_dataset # BigQuery dataset
model: my_table

timeseries: created_at
dimensions:
  - column: region
  - column: category
measures:
  - name: total_revenue
    expression: SUM(revenue)
```

In BigQuery terminology, `database` maps to the **project**, `database_schema` maps to the **dataset**, and `model` maps to the **table**. Measure expressions must use [BigQuery SQL](https://cloud.google.com/bigquery/docs/reference/standard-sql/query-syntax) syntax.

:::note
Rill supports metrics views directly on BigQuery as a live connector. Incremental models and partitioned ingestion are not supported in live connector mode.
:::

### Caching Query Results

By default, dashboard queries against a BigQuery metrics view run live against BigQuery on every interaction. Because BigQuery bills by bytes scanned, repeated dashboard activity can rack up costs quickly. Enable caching to reuse query results between users until the underlying data changes.

Caching is configured under the `cache` block on the metrics view. Because BigQuery is an external/live connector, caching is **off by default** — opt in by setting `cache.key_sql` (or `cache.enabled: true`).

```yaml
type: metrics_view

connector: bigquery
database: my-gcp-project
database_schema: my_dataset
model: my_table

timeseries: created_at
dimensions:
  - column: region
measures:
  - name: total_revenue
    expression: SUM(revenue)

cache:
  key_sql: SELECT MAX(created_at) FROM `my-gcp-project.my_dataset.my_table`
  key_ttl: 5m
```

Rill periodically runs `key_sql` against BigQuery (re-evaluated at most once per `key_ttl`) and uses the returned scalar value as the cache key. When the value changes — for example because a new row landed — the cache is invalidated and the next query repopulates it.

**Pros**

- **Lower BigQuery spend.** Repeat queries (multiple users on the same dashboard, back-and-forth filtering) are served from Rill's cache instead of scanning bytes in BigQuery.
- **Faster dashboards.** Cached results return in milliseconds; no per-query slot allocation or job startup overhead.
- **Pairs with `max_bytes_billed`.** Use caching alongside `max_bytes_billed` (see [Controlling Query Costs](#controlling-query-costs)) to both cap individual query cost and reduce overall query volume.

**Cons**

- **Up to `key_ttl` of staleness.** Between `key_sql` evaluations, new data in BigQuery will not appear on the dashboard. Pick a TTL that matches your data freshness SLA.
- **`key_sql` itself runs on BigQuery.** Make sure it's a cheap query — ideally a `MAX()` on a partitioned/clustered column with a partition filter, or a metadata lookup against `INFORMATION_SCHEMA.PARTITIONS`. A poorly written `key_sql` can scan the whole table on every refresh.
- **Memory usage.** Cached results live in Rill's result cache; very high-cardinality dashboards with many distinct queries will evict older entries.
- **Not suitable for sub-minute freshness.** If you need near-real-time data, leave caching disabled and rely on BigQuery's own result cache instead.
