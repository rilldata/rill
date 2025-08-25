---
title: SQL Models
description: Create models from source data and apply SQL transformations
sidebar_label: SQL Models
sidebar_position: 05
---

## Adding a data model

Add a new data model by either clicking 'model' in the 'Add' menu or select the '...' in any connector view or existing model. When you add a data, a code definition will automatically be created as a `<model_name>.sql` file in the `models` folder of your Rill project.

You can also create a model outside of the application and add it to Rill by placing a `<model_name>.sql` file in the `models` directory containing a DuckDB SQL `SELECT` statement. Rill will automatically detect and parse the model the next time you run `rill start`.


## Annotating your models with properties

In most cases, objects are represented in Rill as YAML files. Models are unique in that any `<model>.sql` file can be considered a model resource in Rill, representing a SQL transformation that you would like to inform using a set of inputs and outputting a view or table (depending on the materialization type). For most other resources, available properties can be set directly via the corresponding YAML file. In the case of a model SQL file though, configurable properties should be set by annotating the top of the file using the following syntax:

```sql
-- @property: value
select * from users
```

We will cover different available configurable properties in the below sections.

### Marking your model SQL file as a model resource type

By default, any new model that is created in a Rill project will populate a corresponding `.sql` file representing the model. Similarly, a `.sql` file that is directly created in the project directory will also be _automatically assumed_ by Rill to be a model by default. Therefore, it is not necessary to annotate the model resource with the `type` property.

For consistency or documentation purposes, if you'd like to annotate your model resource as well with the `type` property, you can do so by adding the following to the top of your `<model_name>.sql`:
```sql
-- @type: model
select * from orders
```

### Specifying the Data Source Connector

The `@connector` annotation explicitly defines which data source connector to use when executing the SQL query. This parameter is essential when working with multiple connectors of the same type, as it determines which connection credentials and source path to utilize. By default, query results are written to the project's default [OLAP engine](/connect/olap#olap-engines).

```sql
-- @connector: bigquery
SELECT * FROM project_id.dataset_name.table_name
```

```sql
-- @connector: postgres
SELECT * FROM mydb.public.users
```

For projects configured with [multiple OLAP engines](/connect/olap/multiple-olap), you can specify both input and output connectors for a SQL model. Nested parameter syntax uses dot notation (e.g., `output.connector`).

```sql
-- @connector: clickhouse
-- @output.connector: clickhouse
select from clickhouse_table
```