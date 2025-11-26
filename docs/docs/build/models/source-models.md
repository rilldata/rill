---
title: Source Models
sidebar_label: Source Models
sidebar_position: 10
---

After [creating a connector to your data source](/build/connectors/data-source), you'll need to create a model to bring that data into Rill. This can be implemented as either a SQL model with [defined connector parameters](/build/models/sql-models#specifying-the-data-source-connector) or as a YAML configuration file. This guide focuses on YAML-based source models, which are auto-generated when using the UI.

```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: snowflake

sql: |
  select * from database_name.schema_name.table_name
```

### YAML Structure

The YAML configuration file contains several key parameters:

- **`type: model`**: Explicitly defines the file type. While Rill automatically detects the file type based on the parent folder, this parameter provides explicit definition.
- **`connector`**: Defines the connector type used to create the model (e.g., `bigquery`, `athena`, `snowflake`, etc.).
- **`sql`**: The actual SQL query to be executed. When nested under `dev:`, the query runs in the Rill Developer environment.
- **`dev`**: Configuration for development mode. Rill Developer runs in dev mode by default, but when deployed to Rill Cloud, the root-level SQL configuration executes. See [Environment Templating](/build/models/templating) for more information.


## Data Explorer

Within Rill, you are able to explore your data before importing into your selcted [OLAP engine](/build/connectors/olap). Navigate to the left lower panel to see the list of [connectors](/build/connectors/data-sources) in your environment. Explore the tables, schema and data preview (up to 150 rows) of your data to ensure you are selecting the correct table.

![release-0 75](<https://cdn.rilldata.com/docs/release-notes/release-075.gif>)
> change


### Import Data Flow

Another surface that you'll see a minimalist version of the data explorer is after connecting to your connector. While this view doesnt allow for data preview, it allows you to view your tables before ingesting into Rill.

:::tip Unsure of the table?

If you're in this flow and not sure you have the correct table selected, close out of the modal and navigate to the left lower panel to view the schema and data preview.

:::

## Retry Configuration

By default, a model will retry if the initial load fails. This helps ensure reliable data processing by automatically retrying failed operations. The default retry settings are:

```yaml
retry:
  attempts: 3 
  delay: 5s
  exponential_backoff: true
```

You can customize the retry behavior to better suit your specific needs. For example, you might want to increase the number of attempts for critical models, adjust the delay between retries, or only retry on specific error types. Use the following configuration in your source YAML:

```yaml
retry:
  attempts: 5
  delay: 10s
  exponential_backoff: true
  if_error_matches:
    - ".*OvercommitTracker.*"
    - ".*Timeout.*"
    - ".*Bad Gateway.*"
```


## Examples

### BigQuery Model
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
  sql: select * from database_name.schema_name.table_name limit 10000

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

Rill provides automatic refresh capabilities for your source models at specified intervals to maintain data currency. This functionality enables you to establish scheduled data ingestion without manual intervention, ensuring your analytics dashboards remain current with the latest information from your data sources. For additional details, see [Scheduled Refreshes](/build/models/data-refresh).

After validating your source model configuration and confirming the data preview appears correct, you can move forward to create your first metrics view. If no additional data transformations are needed, you can choose [Generate Metrics View with AI](/build/metrics-view) from the top-right corner of the interface to initiate Rill's AI-powered dashboard generation.