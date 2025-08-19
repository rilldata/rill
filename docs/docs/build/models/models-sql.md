---
title: SQL Models
description: Create models from source data and apply SQL transformations
sidebar_label: Models (SQL)
sidebar_position: 05
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

### Setting the Connector / OLAP Engine

Another parameter that you can define is the connector parameter. This will indicate which connection credentials to use. This allows you to define multiple connectors to different sources that require unique credentials.
```sql
-- @connector: bigquery
```
