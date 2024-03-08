---
title: Model SQL
sidebar_label: Model SQL
sidebar_position: 20
hide_table_of_contents: true
---

When using Rill Developer, data transformations are powered by DuckDB and their dialect of SQL. Under the hood, by default, data models are created as views in DuckDB. Please check our [modeling](/build/models/models.md) page and [DuckDB documentation](https://duckdb.org/docs/sql/introduction) for more details about how to construct and write your model SQL syntax.

In your Rill project directory, you can also create a `<model_name>.sql` file in the `models` directory containing an appropriate DuckDB `SELECT` statement. Rill will automatically detect and parse the model next time you run `rill start`.

## Model materialization

As mentioned, models will be materialized in DuckDB as views. However, you can choose to materialize them as tables instead of views. To do this, add the following syntax to the **top** of the model SQL file:

```sql
-- @materialize: true
```

Alternatively, it is possible to set it as a [project-wide default](rill-yaml.md#project-wide-defaults) via your `rill.yaml` file:

```yaml
models:
  materialize: true
```

:::info To materialize or not to materialize? 

There are both pros and cons to materializing your models.
- Pros can include improved performance for downstream models and dashboards, especially with the SQL is complex and/or the data size is large. We generally recommend _materializing_ final models that power dashboards (we do this automatically in Rill Cloud).
- Cons can include a degraded keystroke-by-keystroke modeling experience or for specific edge cases, such as when using cross joins.

If unsure, we would generally recommend leaving the defaults and/or [reaching out](contact.md) for further guidance!

:::