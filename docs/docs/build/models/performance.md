---
title: Optimize Your Models
sidebar_label: Performance Optimization
sidebar_position: 45
---

Model performance is critical for maintaining responsive dashboards and ensuring users have access to the most current data. This guide covers strategies for optimizing your Rill models to deliver fast query results while keeping data fresh and up-to-date.

By following these best practices, you can create models that provide both speed and accuracy, delivering insights when your users need them most.

## Local Development / Rill Developer

As discussed in the [templating section](/build/models/templating), there are a few key recommendations to increase model performance in Rill Developer:

1. [Limiting the source data](/build/models/templating#applying-a-one-week-sample-to-the-source-bucket-for-local-development) to a smaller time range (e.g., one week's worth of data instead of the full year)
2. Creating smaller models for development by [applying a raw limit](/build/models/templating#example-conditional-sql-limiting-dev-rows), which will then serve as the starting point for your actual downstream models/modeling
3. [Applying a filter](/build/models/templating#example-leveraging-variables-to-apply-a-filter-and-row-limit-dynamically-to-a-model) to the data to work with only a subset of data

## Model Performance

### Consider which models to materialize

By default, models will be materialized as views (in DuckDB). This allows for a dynamic and highly interactive experience when modeling, such as keystroke-by-keystroke profiling. However, since views are logical in nature, as the complexity and size of your data models continue to grow (especially if the underlying data is very large), this can start to significantly impact performance as these complex queries will need to be continuously re-executed along with a number of profiling queries that the Rill runtime will send in the backend.

In such scenarios, we recommend materializing these models as tables. However, there are some tradeoffs to consider:
- **Pros:** Materializing a model will generally ensure significantly improved performance for downstream dependent models and dashboards.
- **Cons:** Enabling materialization for a model can severely impact or break the "keystroke-by-keystroke" experience and these models may also take longer to update (because the results are being written to a table vs. remaining a view). It can also lead to _degraded_ performance for very specific operations, such as when you need to perform cross joins.

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

## Query Optimization

Query optimization is crucial for maintaining high performance and efficiency, especially when working with data-intensive applications. As Rill dashboards are powered by [OLAP engines](/build/connectors/olap), designed for analytical queries, ensuring that our queries are well-optimized can help maximize the responsiveness and speed of our dashboards. There are also additional potential second-order benefits to optimizing queries in Rill, such as improving ingestion times, how long it takes to build models, how resource-intensive it is to build models, how fast profiling queries run, and more.

### Use appropriate data types and avoid casting when possible

Casting can be expensive, especially when the underlying models are views and not [materialized](#consider-which-models-to-materialize) as a table. For example, if a timestamp column is actually incorrectly typed as a string, then for timeseries charts, Rill ends up having to iterate across each row to try to infer the timestamp and a lot of time parsing has to occur. Similarly, for incorrectly typed or cast columns that are used in calculations, the calculations will have to be constantly looped through, which can be both inefficient and expensive over time (and simply make everything slower).

Similarly, choosing the right data type for each column is also important. Smaller data types, when applicable, consume less memory and can improve query performance. For example, use `INT` instead of `BIGINT` if your data range permits.

### Select the columns you need and avoid `SELECT *` when possible

Because most [OLAP databases](/build/connectors/olap) store data in a columnar format, including [DuckDB](/build/connectors/olap/duckdb), selecting only the columns that you need during the modeling phase ensures that DuckDB will only ingest and store the data _it actually needs_ (speeding up model build times and reducing footprint). Furthermore, columnar formats are optimized for analytical queries, so by selecting only the columns that you need (instead of a blanket `SELECT *`), this will help to minimize data processing times and improve the query execution speed.

### Consider sorting your data by an appropriate timestamp column

Generally speaking, if possible, it is recommended to make sure that your upstream data is relatively well organized and/or sorted by timestamp before being ingested into Rill. This helps to ensure that timeseries queries are efficient when they execute against resulting models and can result in an order of magnitude difference in query performance. This can also help improve the effectiveness of filters by reducing I/O.

:::info When to sort vs. not to sort?

Sorting, especially in DuckDB, _can also be computationally intensive_ and most input data is generally sorted enough (by time). If the data ingested is completely unsorted or sorted by a different, non-timestamp column, it could be worth the computational overhead to sort by a timestamp column (especially if used in a dashboard). If you're unsure, please feel free to [reach out](/contact) and we'd be happy to help you assess the best path forward!

:::

### Use joins efficiently

Plan your joins carefully, especially when working with large datasets. Most [OLAP engines](/build/connectors/olap), DuckDB included, will optimize join operations, but ensuring the join keys are well chosen and considering the size of the datasets being joined can reduce processing time. For example, if you're looking to perform a cross or cartesian join across a very wide table, be sure it's necessary as it can otherwise explode the size of your result set.

### Apply filters early and use WHERE clauses wisely

When possible, it can be good practice to apply filtering early in your queries with `WHERE` clauses to reduce the amount of data being processed in subsequent steps (or downstream models/queries). This can both help to reduce the amount of data being scanned and, given the columnar nature of most [OLAP engines](/build/connectors/olap), significantly speed up queries.

### Optimize your subqueries to leverage joins or CTEs when possible

Subqueries can very often prove to be inefficient and result in suboptimal query execution plans. When possible, it is generally better practice to rewrite subqueries as joins or use Common Table Expressions (CTEs) to make them more readable and potentially more efficient.

### Rather than UNION, consider using UNION ALL when possible

Depending on the [OLAP engine](/build/connectors/olap), `UNION` can be a very expensive operation and much more computationally intensive than `UNION ALL`. For example, when using [DuckDB](/build/connectors/olap/duckdb), a `UNION` will require performing full duplicate eliminations _across all columns_ while a `UNION ALL` will simply concatenate the tables together. If a concatenation is sufficient (for the query), this will be both much quicker and significantly less resource intensive for the query to complete.


## Time Series Transformation

If your time series column is quite granular, this may affect your dashboards as the grain will define how granular the dashboards can be viewed. Instead of handling this in the metrics view by adding a `smallest_time_grain` key, you can use the modeling layer to roll up your data.

:::note Query-time vs Model processing

There are benefits to pre-procesing the data in the model layer but for some quick processing this can be done in the metrics view.

**Query-time processing** (in metrics views):
- Flexible and dynamic
- No storage overhead
- Slightly slower for complex calculations

**Model-level processing** (in SQL models):
- Pre-computed and optimized
- Faster query performance
- Requires model refresh for updates

:::

### DuckDB Time Functions

DuckDB provides a comprehensive toolkit for temporal data manipulation:

- **`DATE_TRUNC`**: Normalize timestamps to consistent intervals (day, week, month, quarter, year)
- **`EXTRACT`**: Extract specific time components (year, quarter, month, day of week, hour)
- **`LAG/LEAD`**: Reference prior or future rows for period-over-period comparisons
- **`DATE_ADD/DATE_SUB`**: Perform date arithmetic for dynamic time ranges
- **`STRFTME`**: Extract strings from a time column.

For comprehensive documentation on all available time functions, see the [DuckDB time functions documentation](https://duckdb.org/docs/stable/sql/functions/timestamp.html).

:::tip not using DuckDB?

Each engine has slightly different functions and syntax for rolling up your data, see your OLAP engine's documentation for more examples.

:::
### Time Aggregation (Roll-ups)

Roll-ups aggregate granular events into coarser intervals. For example, if your data arrives hourly but daily analysis suffices:

```sql
SELECT DATE_TRUNC('day', timestamp_column) AS time_series_column,
        ...  
FROM your_model
```