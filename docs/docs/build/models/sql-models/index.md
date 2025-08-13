---
title: SQL Models
description: Create models from source data and apply SQL transformations
sidebar_label: SQL Models
sidebar_position: 00
---

## Adding a data model

Add a new data model by either clicking 'model' in the 'Add' menu or select the '...' in any connector view or existing model. When you add a data, a code definition will automatically be created as a `<model_name>.sql` file in the `models` folder of your Rill project.

You can also create a model outside of the application and add it to Rill by placing a `<model_name>.sql` file in the `models` directory containing a DuckDB SQL `SELECT` statement. Rill will automatically detect and parse the model the next time you run `rill start`.


## Annotating your models with properties

In most cases, objects are represented in Rill as YAML files. Models are unique in that any `<model>.sql` file can be considered a model resource in Rill, representing a SQL transformation that you would like to inform using a set of inputs and outputting a view or table (depending on the materialization type). For most other resources, available properties can be set directly via the corresponding YAML file. In the case of a model SQL file though, configurable properties should be set by annotating the top of the file using the following syntax:

```sql
-- @property: value
```

We will cover different available configurable properties in the below sections.

### Marking your model SQL file as a model resource type

By default, any new model that is created in a Rill project will populate a corresponding `.sql` file representing the model. Similarly, a `.sql` file that is directly created in the project directory will also be _automatically assumed_ by Rill to be a model by default. Therefore, it is not necessary to annotate the model resource with the `type` property.

For consistency or documentation purposes, if you'd like to annotate your model resource as well with the `type` property, you can do so by adding the following to the top of your `<model_name>.sql`:
```sql
-- @type: model
```

### Setting the Connector 

Another parameter that you can define is the connector parameter. This will indicate which connection credentials to use. This allows you to define multiple connectors to different sources that require unique credentials.
```sql
-- @connector: bigquery
```

## SQL Model Materialization

```sql
-- @materialize: true
```

Model materialization is something to consider when creating intermediate models. Intermediate models are not, by default, materialized and are views in your underlying database engine. There are some pros and cons of enabling it during the development process.

The pros include improved performance for downstream models and dashboards, especially with complex logic and/or large data sizes. Some cons are certain edge cases like cross joins might have a degreaded keystroke-by-keystroke experience, and materialized models are billable.


If you are seeing degraded performance, the first recommendation you'll hear from us is to materialize the metrics powered model.

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
- Pros can include improved performance for downstream models and dashboards, especially with the SQL is complex and/or the data size is large. We generally recommend _materializing_ final models that power dashboards.
- Cons can include a degraded keystroke-by-keystroke modeling experience or for specific edge cases, such as when using cross joins.

If unsure, we would generally recommend leaving the defaults and/or [reaching out](/contact) for further guidance!

:::

### Materializing the model as a table and then utilizing the `ref` function

If you <u>need</u> to use the `PIVOT` statement specifically but don't want to specify an `IN` filter, then you will need to inform Rill to materialize this model as a table **and** leverage the `ref` function as well (for proper DAG resolution). Using the same example, this would instead look something like:

```sql
-- @materialize: true

PIVOT {{ ref "table_name" }} ON column_name USING SUM(measure)
```
