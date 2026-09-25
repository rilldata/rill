---
title: Query Attributes
description: Tag the queries a metrics view sends to your OLAP engine for cost attribution and auditing
sidebar_label: Query Attributes
sidebar_position: 30
---

Query attributes are key-value pairs that Rill attaches to every query a metrics view sends to its OLAP engine. Use them to attribute warehouse cost to users or teams, or to audit who ran which query, from your engine's own query history.

Query attributes apply to every query Rill runs against the metrics view: explore and canvas dashboards, alerts, reports, APIs, and AI chat. They are delivered only to engines that support them; see [Delivery by OLAP engine](#delivery-by-olap-engine).

## Configuring Query Attributes

Add a `query_attributes` map to your metrics view YAML. Values are strings and support [templating](/developers/build/connectors/templating) with [user attributes](/developers/build/metrics-view/security#user-attributes) and environment variables:

```yaml
# metrics/orders.yaml
type: metrics_view
model: orders
timeseries: order_date

dimensions:
  - column: region
measures:
  - name: total_revenue
    expression: SUM(revenue)

query_attributes:
  rill_user: '{{ .user.email }}'
  rill_department: '{{ .user.department | default "unknown" }}'
  rill_project: '{{ .env.project_name }}'
```

Rill resolves the templates for each query, using the attributes of the user who issued it. With the example above, a query from `jane@example.com` in the `sales` department is sent with `rill_user = jane@example.com` and `rill_department = sales`.

The following rules apply:

- Keys may contain only letters, digits, underscores (`_`), hyphens (`-`), and dots (`.`). Rill reports any other key as a parse error on the metrics view.
- A template that fails to resolve, for example because it references an environment variable that is not defined, fails the query.
- A user attribute that is not set resolves to the literal text `<no value>`. Use `default` to provide a fallback, as in `rill_department` above. This also applies when Rill validates the metrics view, which runs without a user.

:::tip
Custom attributes passed from an [embedded dashboard](/developers/embed/iframe) are available as `.user.<attribute>`, so you can tag queries with your application's own tenant or customer ID.
:::

## Delivery by OLAP Engine

Each OLAP engine receives query attributes in its own way, and only some engines receive them at all:

| OLAP engine | How attributes are sent | Where to find them |
|---|---|---|
| ClickHouse | Appended to the query as `SETTINGS key = 'value'` | The query text and the `Settings` column of `system.query_log` |
| Druid | Added to the [query context](https://druid.apache.org/docs/latest/querying/query-context/) | Druid request logs |
| Databricks | Sent as [query tags](#databricks-query-tags) | The `query_tags` column of `system.query.history` |
| DuckDB, MotherDuck, DuckLake, Snowflake, BigQuery, Pinot, StarRocks, Redshift, Athena, Postgres, MySQL | Not sent | Not applicable |

### ClickHouse

ClickHouse only accepts settings it knows about. A key that is not a built-in ClickHouse setting must start with a prefix listed in the server's [`custom_settings_prefixes`](https://clickhouse.com/docs/operations/server-configuration-parameters/settings#custom_settings_prefixes) configuration, such as `custom_`. Otherwise, ClickHouse rejects the query.

```yaml
query_attributes:
  custom_rill_user: '{{ .user.email }}'
```

If the ClickHouse user Rill connects with is in read-only mode and cannot change settings, Rill does not send query attributes.

### Druid

Rill merges query attributes into the Druid query context. Keys that Rill sets itself (`sqlQueryId`, `enableTimeBoundaryPlanning`, `useCache`, `populateCache`, and `priority`) cannot be overridden and are skipped.

### Databricks Query Tags

Rill sends query attributes to Databricks as [query tags](https://docs.databricks.com/aws/en/sql/user/queries/query-tags), attached to each statement. They are recorded in the `query_tags` column (a `map<string, string>`) of the [`system.query.history`](https://docs.databricks.com/aws/en/admin/system-tables/query-history) system table, which you can query to break down SQL warehouse usage:

```sql
SELECT
  query_tags['rill_user'] AS rill_user,
  COUNT(*) AS queries,
  SUM(total_duration_ms) AS total_duration_ms
FROM system.query.history
WHERE query_tags['rill_user'] IS NOT NULL
GROUP BY 1
ORDER BY total_duration_ms DESC
```

Databricks applies its own limits to query tags: at most 20 tags per query, keys and values of at most 128 characters, and keys that do not contain `,`, `:`, `-`, `/`, `=`, or `.`. Use underscores in your keys, as in the examples on this page.

Query tags are a Databricks Public Preview feature. If your workspace does not support them, Databricks rejects every query that carries tags, and the metrics view fails to reconcile with this error:

```
[CONFIG_NOT_AVAILABLE.WITHOUT_SUGGESTION] Configuration query_tags is not available.
```

Rill does not retry the query without tags. To fix the error, remove `query_attributes` from the metrics view, or contact your Databricks account team to get access to the query tags preview for your workspace.

:::note
Query tags are sent only over the default Thrift protocol. When the connector uses the Statement Execution API (SEA) backend, which Rill selects automatically for [Lakehouse//RT warehouses](/developers/build/connectors/data-source/databricks#manual-configuration) or when you set `use_kernel: true`, query attributes are not sent.
:::

:::info
Refer to the [`query_attributes` property](/reference/project-files/metrics-views#query_attributes) in the metrics view YAML reference.
:::
