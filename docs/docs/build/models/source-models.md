---
title: Source Models
sidebar_label: Source Models
sidebar_position: 10
---

After [creating a connector to your data source](/connect/data-source), you'll need to create a model to bring that data into Rill. This can be implemented as either a SQL model with [defined connector parameters](/build/models/sql-models#specifying-the-data-source-connector) or as a YAML configuration file. This guide focuses on YAML-based source models.

## Overview

Once you can see your tables through the connector, you can directly create a Rill model and ingest the source data. Rill includes built-in safeguards to prevent excessive costs and time consumption during the initial data read. These safeguards are configured through the `dev:` parameter settings. For more information on dev/prod configurations in models, see [Dev/Prod Environments](/build/models/templating).

```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: "duckdb"

dev:
  sql: |
    select * from read_csv('gs://my-bucket/path/to/file.csv', auto_detect=true, ignore_errors=1, header=true) limit 10000

sql: |
  select * from read_csv('gs://my-bucket/path/to/file.csv', auto_detect=true, ignore_errors=1, header=true)
```

## Source Model Configuration

### YAML Structure

The YAML configuration file contains several key parameters:

- **`type: model`**: Explicitly defines the file type. While Rill automatically detects the file type based on the parent folder, this parameter provides explicit definition.
- **`connector`**: Defines the connector type used to create the model (e.g., `bigquery`, `athena`, `snowflake`, etc.).
- **`sql`**: The actual SQL query to be executed. When nested under `dev:`, the query runs in Rill Developer environment.
- **`dev`**: Configuration for development mode. Rill Developer runs in dev mode by default, but when deployed to Rill Cloud, the root-level SQL configuration executes.

### Automatic Refresh Schedule

Rill can automatically refresh your source models at specified intervals to ensure your data stays current. This feature allows you to set up scheduled data ingestion without manual intervention, keeping your analytics dashboards up-to-date with the latest information from your data sources.


```yaml
refresh:
  every: 24h
```

For more information, see [Scheduled Refreshes](/build/models/data-refresh).


## Examples

### Big Query Model
```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: bigquery

dev:
  sql: select * from project_id.dataset_name.table_name limit 10000

sql: select * from project_id.dataset_name.table_name

```

### Snowflake Model
```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: "snowflake"

dev:
  sql: select * from database_name.schema_name.table_namelimit 10000

sql: select * from database_name.schema_name.table_name

```


### S3 Model
```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: "duckdb"

dev:
  sql: |
    select * from read_csv('s3://my-bucket/path/to/file.csv', auto_detect=true, ignore_errors=1, header=true) limit 10000

sql: |
  select * from read_csv('s3://my-bucket/path/to/file.csv', auto_detect=true, ignore_errors=1, header=true)
```

For more information, see our [model reference documentation](/reference/project-files/models)!

## Next Steps

Once you've validated your source model configuration and confirmed the data preview looks correct, you can proceed to create your first metrics view. If no additional data transformations are required, you can select [**Generate Metrics View with AI**](/build/metrics-view) from the top-right corner of the interface. This will launch Rill's AI-powered dashboard generation to help you get started with your analytics journey.