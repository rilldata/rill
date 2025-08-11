---
note: GENERATED. DO NOT EDIT.
title: Model SQL
sidebar_position: 33
---

When using Rill Developer, data transformations are powered by DuckDB and their dialect of SQL. Under the hood, by default, data models are created as views in DuckDB. Please check our modeling page and DuckDB documentation for more details about how to construct and write your model SQL syntax.

In your Rill project directory, you can also create a `<model_name>.sql` file containing an appropriate DuckDB `SELECT` statement, most commonly within the default `models` directory, to represent a model (or set of SQL transformations). Rill will automatically detect and parse the model next time you run `rill start`.

  ### Annotating your models with properties
  In most cases, objects are represented in Rill as YAML files. Models are unique in that any model.sql file can be considered a model resource in Rill, representing a SQL transformation that you would like to inform using a set of inputs and outputting a view or table (depending on the materialization type). For most other resources, available properties can be set directly via the corresponding YAML file. In the case of a model SQL file though, configurable properties should be set by annotating the top of the file using the following syntax:
  ```sql
  -- @property: value
  ```
  We will cover different available configurable properties in the below sections.


## Properties

### `type`

_[string]_ - By default, any new model that is created in a Rill project will populate a corresponding .sql file representing the model. Similarly, a .sql file that is directly created in the project directory will also be automatically assumed by Rill to be a model by default. Therefore, it is not necessary to annotate the model resource with the type property.

For consistency or documentation purposes, if you'd like to annotate your model resource as well with the type property, you can do so by adding the following to the top of your model_name.sql:
```sql
-- @type: model
```
 

### `materialize`

_[boolean]_ - As mentioned, models will be materialized in DuckDB as views by default. However, you can choose to materialize them as tables instead of views. To do this, you can add the following annotation to the top of your model SQL file:
```sql
-- @materialize: true
```

Alternatively, it is possible to set it as a project-wide default as well that your models inherit via your rill.yaml file:
```yaml
models:
  materialize: true
```

:::info To materialize or not to materialize?

There are both pros and cons to materializing your models.
- Pros can include improved performance for downstream models and dashboards, especially with the SQL is complex and/or the data size is large. We generally recommend _materializing_ final models that power dashboards.
- Cons can include a degraded keystroke-by-keystroke modeling experience or for specific edge cases, such as when using cross joins.

If unsure, we would generally recommend leaving the defaults and/or reaching out for further guidance!
:::
 

## Examples

### Materialized SQL Model

```sql-
-- Model SQL
-- Reference  documentation: https://docs.rilldata.com/reference/project-files/models
-- @type: model
-- @materialize: true

select * from your_table
```
