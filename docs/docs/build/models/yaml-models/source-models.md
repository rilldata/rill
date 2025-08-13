---
title: Source Models
sidebar_label: Source Models
sidebar_position: 00
---

After creating a connector to your data source, you'll need to create a model to bring that data into Rill. This can be implemented as either a SQL model with [defined connector parameters](/build/models/sql-models#setting-the-connector) or as a YAML configuration file. This guide focuses on YAML-based source models.

## Overview

Once you can see your tables through the connector, you can directly create a Rill model and ingest the source data. Rill includes built-in safeguards to prevent excessive costs and time consumption during the initial data read. These safeguards are configured through the `dev:` partition settings. For more information on dev/prod configurations in models, see [Dev/Prod Environments](/build/models/templating).

<img src="/img/deploy/templating/gcs-env-example.png" class="rounded-gif" alt="GCS Environment Configuration Example" />
<br />

## Source Model Configuration

### YAML Structure

The YAML configuration file contains several key parameters:

- **`type: model`**: Explicitly defines the file type. While Rill automatically detects the file type based on the parent folder, this parameter provides explicit definition.
- **`connector`**: Automatically populated based on the connector type used to create the model (e.g., `bigquery`, `athena`, `snowflake`, etc.).
- **`dev`**: Configuration for development mode. Rill Developer runs in dev mode by default, but when deployed to Rill Cloud, the root-level SQL configuration executes.
- **`sql`**: The actual SQL query to be executed. When nested under `dev:`, the query runs in Rill Developer environment.

## Data Preview and Validation

### Table Preview

Rill automatically generates a preview of your data (first 150 rows) to help verify that the output table structure is correct and identify any potential issues that need to be addressed in the SQL configuration, such as data type detection problems.

### Schema Details

The left panel displays comprehensive information about your dataset and column contents:

- **Dataset Overview**: Total row and column counts
- **Data Quality Metrics**: Number of dropped rows and columns
- **Column Analysis**: 
  - Column names and data types
  - Distinct value counts for string columns
  - Basic numeric statistics (minimum, maximum, median, etc.)

This information helps you validate your model configuration and ensure data quality before proceeding with the full data ingestion.


## Next Steps

Once you've validated your source model configuration and confirmed the data preview looks correct, you can proceed to create your first metrics view. If no additional data transformations are required, you can select [**Generate Metrics View with AI**](/build/metrics-view) from the top-right corner of the interface. This will launch Rill's AI-powered dashboard generation to help you get started with your analytics journey.