---
title: Optimize Your Models
sidebar_label: Performance Optimization
sidebar_position: 45
---

Model performance is critical for maintaining responsive dashboards and ensuring users have access to the most current data. This guide covers strategies for optimizing your Rill models to deliver fast query results while keeping data fresh and up-to-date.

By following these best practices, you can create models that provide both speed and accuracy, delivering insights when your users need them most.

:::info Working with very large data from the get go?

Generally speaking, Rill's [embedded DuckDB OLAP engine](/connect/olap/duckdb) works very well out-of-the-box for datasets _up to around 50GB in size_. If you plan to be working with and ingesting volumes of data larger than 50GB, please [**get in touch**](/contact) and we can explore using one of our other enterprise-grade [OLAP engine](/connect/olap) options. 

:::


## Local Development / Rill Developer

When developing your project in Rill Developer, we highly recommend setting up dev partitions to work with a subset of data. This approach allows you to iterate quickly on model changes while maintaining fast development cycles.

### Work with a subset of your source data for local development and modeling

As a general rule of thumb, we strongly recommend working with a segment of the data for modeling purposes as part of your local development workflow. This becomes increasingly important as the size of your source data grows in size, which will help ensure that your developer experience remains optimal in Rill Developer. With Rill Developer, it's best to work with a "dev partition" to help validate that the model logic is correct and producing results as expected. Then, once finalized, you can push to Rill Cloud for primary dashboard consumption (including analysis and sharing as necessary).

There are a few ways to achieve this:
1. Limiting the source data to a smaller time range (e.g. one week's worth of data instead of the full year)
2. Creating intermediate staging models from your source data, either through statistical sampling or by applying a raw limit, which will then serve as the starting point for your actual downstream models / modeling
3. Use data from a dev / staging environment

### Limiting the source data to a smaller time range

There are different ways this can be achieved and the method also depends heavily on the data source being used. For example, assuming we had a [Big Query source](/connect/data-source/bigquery) created from the connectors UI.
```yaml
type: model
connector: bigquery
materialize: true
sql: select * from project_id.dataset.table
dev:
  sql: select * from project_id.dataset.table limit 10000
```
Instead of a general hard limit of 10000, which may give a subpar experience in the dashboard review, we can change the YAML to include a dev partitino that filters on a time column instead.


```yaml
type: model
connector: bigquery
materialize: true
sql: select * from project_id.dataset.table
dev:
  sql: select * from project_id.dataset.table {{ if dev }} where timestamp_column >= '2024-12-01' AND timestamp_column < '2025-01-01' {{ end }}
```

### Creating intermediate staging models

Another option would be to create intermediate staging models from your sources, either through [statistical sampling](https://duckdb.org/docs/sql/samples.html) or by applying a [raw limit](https://duckdb.org/docs/sql/query_syntax/limit.html), to reduce the size of your models in development. For example, with [templating](/connect/templating) and [environments](/build/models/templating.md), this `model.sql` applies a five percent sample to a source:

```sql
-- @materialize

SELECT * FROM {{ ref "source_name" }}
{{ if dev }} USING SAMPLE 5% {{ end }}
```

Similarly, if we were less concerned about skewing our sample data, we could apply a simple `LIMIT` to our source data:

```sql
-- @materialize

SELECT * FROM {{ ref "source_name" }}
{{ if dev }} LIMIT 1000 {{ end }}
```

:::info Why are we materializing the intermediate model?

As mentioned in a [Model Performance](#consider-which-models-to-materialize), models by default will be materialized as _views_ in DuckDB unless otherwise specified. When working with intermediate models, we'll want them to be stored as **tables** to ensure downstream models and queries remain performant.
:::

## Model Performance

### Consider which models to materialize

By default, models will be materialized as views (in DuckDB). This allows for a dynamic and highly interactive experience when modeling, such as keystroke-by-keystroke profiling. However, since views are logical in nature, as the complexity and size of your data models continue grow (especially if the underlying data is very large), this can start to significantly impact performance as these complex queries will need to be continuously re-executed along with a number of profiling queries that the Rill runtime will send in the backend. 

In such scenarios, we recommend materializing these models as tables. However, there are some tradeoffs to consider.
- **Pros:** Materializing a model will generally ensure significantly improved performance for downstream dependent models and dashboards. 
- **Cons:** Enabling materialization for a model can severely impact or break the "keystroke-by-keystroke" experience and these models may also take longer to update (because the results are being written to a table vs remaining a view). It can also lead to _degraded_ performance for very specific operations, such as when you need to perform cross joins.

:::tip Materialize models powering dashboards

We strongly recommend materializing final models that are being used directly in dashboards to ensure this data is served more quickly. 

:::


### Materialization

Model materialization is something to consider when creating intermediate models. Other than [source models](/build/models/source-models), intermediate models are not, by default, materialized and are views in your underlying database engine. There are some pros and cons to enabling it during the development process.

```sql
-- model.sql
-- @materialize: true
```

```yaml
# model.yaml
materialize: true
```

The pros include improved performance for downstream models and dashboards, especially with complex logic and/or large data sizes. Some cons are certain edge cases like cross joins might have a degraded keystroke-by-keystroke experience, and materialized models are billable.

If you are seeing degraded performance, the first recommendation you'll hear from us is to materialize the metrics-powered model.

### Default Model Materialization

If you want, you can change the default behavior of all models in Rill by setting the default model behavior in the rill.yaml file.

```yaml
models:
  materialize: true
```

To override this on a per-model basis, simply set the specific model.sql to false.

```sql
-- @materialize: false
```

:::info To materialize or not to materialize?

There are both pros and cons to materializing your models.
- Pros can include improved performance for downstream models and dashboards, especially when the SQL is complex and/or the data size is large. We generally recommend _materializing_ final models that power dashboards.
- Cons can include a degraded keystroke-by-keystroke modeling experience or for specific edge cases, such as when using cross joins.

If unsure, we would generally recommend leaving the defaults and/or [reaching out](/contact) for further guidance!

:::




## Refreshing Source

Another area to review when your data source starts getting larger is the ingestion performance. By default, when refreshing a [source model](/build/models/source-models) in Rill, it drops and re-ingests the entire table/file. When your data is small, this isn't an issue but is not appropriate for larger datasets. In these cases, we recommend using [partitions and incremental models](/build/models/incremental-partitioned-models).

### Partitioned Models

Partitioned models divide your data into logical segments based on specific criteria, typically time-based columns like dates. This approach allows you to selectively refresh a partition where you know data has been altered.


Example partition configuration:
```yaml
partitions:
    glob:
      path: 'gs://my-bucket/**/*.parquet'
```

### Incremental Models

Incremental models only process new or changed data since the last refresh, rather than reprocessing the entire dataset. This dramatically improves performance for large datasets:

- **Faster Refresh Times**: Process only delta changes instead of full datasets
- **Reduced Resource Usage**: Lower CPU, memory, and storage requirements
- **Frequent Updates**: Enable near real-time data updates without performance degradation
- **Cost Efficiency**: Minimize compute costs for large-scale data processing

```yaml
type: model
incremental: true

partitions:
  glob:
    path: gs://rilldata-public/github-analytics/Clickhouse/2024/*/*
    partition: directory
  
sql: |
  SELECT * 
     FROM read_parquet('{{ .partition.uri }}/commits_*.parquet') 
    WHERE '{{ .partition.uri }}' IS NOT NULL
```

By combining partitioning and incremental processing, you'll significantly reduce model refresh times and ensure your dashboards display the most current information.
## Query Optimization

Query optimization is crucial for maintaining high performance and efficiency, especially when working with data-intensive applications. As Rill dashboards are powered by [OLAP engines](/connect/olap), designed for analytical queries, ensuring that our queries are well-optimized can help maximize the responsiveness and speed of our dashboards. There are also additional potential second-order benefits to optimizing queries in Rill, such as improving ingestion times, how long it takes to build models, how resource intensive it is to build models, how fast profiling queries run, and more. 

### Use appropriate data types and avoid casting when possible

Casting can be expensive, especially when the underlying models are views and not [materialized](#consider-which-models-to-materialize) as a table. For example, if a timestamp column is actually incorrectly typed as a string, then for timeseries charts, Rill ends up having to iterate across each row to try to infer the timestamp and a lot of time parsing has to occur. Similarly, for incorrectly typed or casted columns that are used in calculations, the calculations will have to be constantly looped through, which can be both inefficient and expensive over time (and simply make everything slower). 

Similarly, choosing the right data type for each column is also important. Smaller data types, when applicable, consume less memory and can improve query performance. For example, use `INT` instead of `BIGINT` if your data range permits.

### Select the columns you need and avoid `SELECT *` when possible

Because most [OLAP databases](/connect/olap) store data in a columnar format, including [DuckDB](/connect/olap/duckdb), selecting only the columns that you need during the modeling phase ensures that DuckDB will only ingest and store the data _it actually needs_ (speeding up model build times and reducing footprint). Furthermore, columnar formats are optimized for analytical queries so by selecting only the columns that you need (instead of a blanket `SELECT *`), this will help to minimize data processing times and improve the query execution speed.

### Consider sorting your data by an appropriate timestamp column

Generally speaking, if possible, it is recommended to make sure that your upstream data is relatively well organized and/or sorted by timestamp before being ingested into Rill. This helps to ensure that timeseries queries are efficient when they execute against resulting models and can result in an order of magnitude difference in query performance. This can also help improve the effectiveness of filters by reducing IO.

:::info When to sort vs not to sort?

Sorting, especially in DuckDB, _can also be computationally intensive_ and most input data is generally sorted enough (by time). If the data ingested is completely unsorted or sorted by a different, non-timestamp column, it could be worth the computational overhead to sort by a timestamp column (especially if used in a dashboard). If you're unsure, please feel free to [reach out](/contact) and we'd be happy to help you assess the best path forward!

:::

### Use joins efficiently

Plan your joins carefully, especially when working with large datasets. Most [OLAP engines](/connect/olap), DuckDB included, will optimize join operations, but ensuring the join keys are well chosen and considering the size of the datasets being joined can reduce processing time. For example, if you're looking to perform a cross or cartesian join across a very wide table, be sure it's necessary as it can otherwise explode the size of your result set. 

### Apply filters early and use WHERE clauses wisely

When possible, it can be good practice to apply filtering early in your queries with `WHERE` clauses to reduce the amount of data being processed in subsequent steps (or downstream models / queries). This can both help to reduce the amount of data being scanned and, given the columnar nature of most [OLAP engines](/connect/olap), significantly speed up queries.

### Optimize your subqueries to leverage joins or CTEs when possible

Subqueries can very often prove to be inefficient and result in suboptimal query execution plans. When possible, it is generally better practice to rewrite subqueries as joins or use Common Table Expressions (CTEs) to make them more readable and potentially more efficient.


### Rather than UNION, consider using UNION ALL when possible

Depending on the [OLAP engine](/connect/olap), `UNION` can be a very expensive operation and much more computationally intensive than `UNION ALL`. For example, when using [DuckDB](/connect/olap/duckdb), a `UNION` will require performing full duplicate eliminations _across all columns_ while a `UNION ALL` will simply concatenate the tables together. If a concatenation is sufficient (for the query), this will be both much quicker and significantly less resource intensive for the query to complete.


## OLAP Engines

Often, data in external [OLAP engines](/connect/olap) is quite large in size and requires different considerations to improve performance. At Rill, we manage clusters in the 100's of TB so included some tips below based on our experience.

### Data Lifecycle Management 

One common way to decrease overall data size and improve query performance (by scanning less data) is to rollup your data to higher time grains historically. Typically, this means taking hourly data and rollup up to daily data when the additional level of granularity is no longer necessary for business needs. Databases like Apache Druid have these lifecycle tools built in or reach out to Rill with questions.

A couple of considerations when rolling data from lower to higher time grains:

- Daily data loses time zone querying as everything is rolled up to a single time zone (usually UTC)
- Consider hashed compaction when going from hourly to daily to reduce data size even further
- Watch out for rolling up metrics. Some metrics should be summed - but others (like a bid floor or campaign budget) should stay unique and be rolled up as a max

### Dimension Stripping

Dimension stripping is another tool to reduce data size by removing high cardinality fields that are not required for analysis. While this can be done upfront in the data set, another practice would be to drop these fields at certain intervals when they no longer add business value. Most frequently, we see a couple decision points where these fields are dropped:

- After a day to first week, dropping user level details no longer needed for monitoring
- After a week to multiple weeks, dropping "double click" level details that aren't needed for reporting (e.g. the minor release number on an Operating System field)
- After a month to months, dropping fields no longer interesting for analysis

### Sampling & Datasketches

There are times where you may look at sampling data feeds to trade data accuracy for lower costs and faster query speeds. Sampling involves sending only a percentage of your data, then extrapolating the values to get an estimate. Rill does not recommend sampling your primary KPIs, any records that require a join or are tied to revenue. This filtered data should be decided in random fashion to not skew or bias the results. Please note, tracking uniques is not recommended if you choose to sample. 

If looking to track uniques, but with smaller datasets and significantly improved performance, you can load unique values (ip addresses, user ids, URLs, etc) with [datasketches](https://datasketches.apache.org). There are multiple types of datasketches supported depending on your engine. At a high level, datasketches use algorithms to approximate unique values. Common use cases for datasketches include count distincts (campaign reach, unique visitors) and quantiles (time spent, frequency). Check out the [Apache Datasketches](https://datasketches.apache.org/docs/Architecture/MajorSketchFamilies.html) site for more details on methodology and use cases.

### Lookups

While joins can kill the performance of [OLAP engines](/connect/olap), lookups (key-value pairs) are common to reduce data size and improve query speeds. Lookups can be done during ingestion time (a static lookup to enrich the source data) or at query time (dynamic lookups).

**Static Lookups** 

Static Lookups are lookups that are ingested at processing time. When a record is being processed, if a match is found between the record and lookup's key, the lookup's corresponding value at that moment in time is extracted and carbon-copied into Druid for the records it processed.

Static lookups are best suited for:

- Dimensions with values that require a historical record for how it has changed over time
- Values are never expected to change (leverage Dynamic Lookups if the values are expected to change)
- Extremely large lookups (hundreds of thousands or records or >50MB lookup file) to improve query performance

Customers typically store lookup values in s3 or GCS and the lookup file is then updated by customers as needed and consumed by ETL logic.

**Dynamic Lookups**

Since static lookups transform and store the data permanently, any changes to the mapping would require reprocessing the entire data set to ensure consistency. To address the case when values in a lookup are expected to change with time we, developed dynamic lookups. Dynamic Lookups, also known as Query Time Lookups, are lookups that retrieved at query time, as apposed to being used at ingestion time.

Benefits of dynamic lookups include:

- Historical continuity for dimensions that change frequently without reprocessing the entire data set
- Time savings, because there is no data set reprocessing required to complete the update
- Dynamic lookups are kept separate from the data set. Thus, any human errors introduced in the lookup do not impact the underlying data set
- Ability for users to create new dimension tables from metadata associated with a dimension table. For example, account ownership can change during the course of a quarter. In such cases, a dynamic lookup can be updated on the fly to reflect the most current changes