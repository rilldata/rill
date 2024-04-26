---
title: Model SQL
sidebar_label: Model SQL
sidebar_position: 20
hide_table_of_contents: true
---

When using Rill Developer, data transformations are powered by DuckDB and their dialect of SQL. Under the hood, _by default_, data models are created as views in DuckDB. Please check our [modeling](/build/models/models.md) page and [DuckDB documentation](https://duckdb.org/docs/sql/introduction) for more details about how to construct and write your model SQL syntax.

In your Rill project directory, you can also create a `<model_name>.sql` file containing an appropriate DuckDB `SELECT` statement, most commonly within the default `models` directory, to represent a model (or set of transformations). Rill will automatically detect and parse the model next time you run `rill start`.

:::tip Did you know?

Rill will automatically assume any `.sql` file within the Rill project directory to be a model, including `.sql` files that might be nested _under multiple levels_ or within subfolders in a directory (such as `models`). Models are unique in that they are a resource that doesn't necessarily need the `kind` property specified.

:::

## Annotating your models with properties

In most cases, objects are represented in Rill as YAML files. Models are unique in that any `<model>.sql` file can be considered a model object in Rill, representing a SQL transformation that you would like to inform using a set of inputs and outputting a view or table (depending on the materialization type). For most other objects, available properties can be set directly via the YAML file. In the case of a model SQL file though, configurable properties can be set by annotating the top of the file using the following syntax:

```sql
-- @property: value
```

We will cover different available configurable properties in the below sections.

## Marking your model SQL file as a model resource

By default, any new model that is created in a Rill project will populate a corresponding `.sql` file representing the model. Similarly, a `.sql` file that is directly created in the project directory will also be _assumed_ by Rill to be a model by default. Therefore, it is not necessary to annotate the model resource with the `kind` property.

For consistency or documentation purposes, if you'd like to annotate your model resource as well with the `kind` property, you can do so by adding the following to the top of your `<model_name>.sql`:
```sql
-- @kind: model
```

:::note

This only applies to models as models are defined using _SQL files_. For other resources whose configuration is handled in YAML, the `kind` property is still required.

:::

## Model materialization

As mentioned, models will be materialized in DuckDB as views by default. However, you can choose to materialize them as tables instead of views. To do this, you can add the following annotation to the top of your model SQL file:

```sql
-- @materialize: true
```

Alternatively, it is possible to set it as a [project-wide default](rill-yaml.md#project-wide-defaults) as well that your models inherit via your `rill.yaml` file:

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