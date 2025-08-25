---
title: Incremental Models 
description: Create Incremental Models
sidebar_label: Incremental Models
sidebar_position: 20
---

Incremental models help with the ingestion of large datasets by allowing a dataset to be broken down into smaller sections for ingestion, rather than reading the entire dataset at once. Unlike [standard SQL models](/build/models/sql-models) that are created via a .sql file, incremental models are defined in a YAML file and are used when a large dataset needs to be incrementally ingested to improve ingestion costs and time.

:::note Take a look at the Reference!

If you are unsure about the required parameters, please review the [reference page for Advanced Models](/reference/project-files/models).

:::

Rill supports incremental models on either cloud storage or data warehouses, but the parameters to set these up will be different. Cloud storage requires the `glob` parameter while data warehouses will need to use `sql`.

See [our reference documentation](/reference/project-files/models) for more information.

:::tip Need help setting up Incremental Models?

Please [reach out to us](/contact) if you have any questions regarding incremental modeling!

:::

## Creating an Incremental Model

In order to enable an incremental model, you will need to set the following: `incremental: true`.

```yaml
type: model
incremental: true

sql: # some SQL query from source_table
```

:::warning Duplicate Data

Incremental models default to an append strategy, and with neither `state` nor `partition` defined, your data will append data per incremental refresh from the source table. This will result in duplicate data and is not recommended. Instead, use the `merge_strategy` with a `unique_key` to ensure duplicate data is not ingested.

:::

:::warning Late Arriving Data

If you have late arriving data, you will need to keep this in mind when designing your incremental model. If you simply use max(date) from the source, you may risk leaving out late arriving data. Depending on your specific use case, you might consider a larger time difference and use a `merge` as your `incremental_strategy`.

:::

### Incremental Models with State Defined (Optional)

If your data is not [partitioned](/build/models/partitioned-models), you can define the incremental model with a predefined `state` parameter. This is only useful for multi-connector incremental ingestion such as BigQuery to DuckDB.

```yaml
type: model
incremental: true
connector: bigquery

state:
  sql: SELECT MAX(date) as max_date

sql: |
      SELECT ... FROM events
        {{ if incremental }}
            WHERE event_time > '{{.state.max_date}}'
        {{end}}
output:
  connector: duckdb
```

Once state is defined in an incremental model, its value can be used as a variable in your SQL statement. In the above example, the state gets the most recent date from the model and when incrementally refreshing, ingests data for events that are more recent than the state's max_date.

:::tip

You can verify the current value of your state in the left-hand panel under Incremental Processing.

:::

### Refreshing an Incremental Model

When you are testing with incremental models in Rill Developer, you will notice a change in the refresh functionality. Instead of a full refresh, you are given the option for `incremental refresh`.

<img src='/img/tutorials/advanced-models/now-incremental.png' class='rounded-gif' />
<br />

:::tip What's the difference?

Once increments are enabled on a model, this grants you the ability to refresh the model in increments, instead of loading the full data each time. This is handy when your data is massive and re-ingesting the data may take time. For a project in production, this allows for less downtime when needing to update your dashboards when the source data is updated.

There are times when a full refresh may be required. In these cases, running the full refresh is equivalent to running a normal refresh with incremental disabled.

:::

When selecting to refresh incrementally, what is being run in the CLI is:

```bash
rill project refresh --local --model <model_name>
```

Keep in mind that if you select `Full refresh`, this will start the ingestion of **all of your data** from scratch. Only use this when absolutely required. When running a full refresh, the CLI command is:

```bash
rill project refresh --local --model <model_name> --full
```

## Model Change Modes

Configure how changes to your model specifications are applied:

```yaml
# model.yaml
change_mode: reset  # Options: reset (default), manual, patch
```

- `reset`: changing the model automatically leads to a full refresh (default)
- `manual`: changing the model stops refreshes until a manual incremental or full refresh is run
- `patch`: changing the model automatically changes to the new logic without a reset (only works for models with `incremental: true`)
