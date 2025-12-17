---
title: Python Scripts
description: Connect to data via Python Scripts
sidebar_label: Python
sidebar_position: 52
---

Use Python scripts to extract, transform, and load data from various sources into your Rill project. Python scripts provide flexibility for connecting to APIs, databases, or other data sources that may not have direct connector support.

## Overview

The Python Scripts connector allows you to run Python scripts that extract data and output it in formats that Rill can consume (CSV, JSON, Parquet). This is useful for:

- **Custom API integrations**: Connect to APIs that don't have native Rill connectors
- **Complex data transformations**: Perform advanced data processing before ingestion
- **Data pipeline orchestration**: Use tools like [dlt (data load tool)](https://dlthub.com/) to build robust data pipelines
- **Legacy system integration**: Extract data from systems that only expose data through custom interfaces

## Using Python Scripts

### Basic Setup

Create a Python script that outputs data in a format Rill can consume. The script should write data to a file that your Rill project can access.

**Example Python script (`scripts/extract_data.py`):**

```python
import csv
import requests

# Extract data from an API
response = requests.get('https://api.example.com/data')
data = response.json()

# Transform and write to CSV
with open('data/extracted_data.csv', 'w', newline='') as f:
    writer = csv.DictWriter(f, fieldnames=data[0].keys())
    writer.writeheader()
    writer.writerows(data)
```

**Reference the output file in your Rill model:**

```yaml
type: model
materialize: true

connector: duckdb

sql: |
  select * from read_csv('data/extracted_data.csv', auto_detect=true, ignore_errors=1, header=true)
```

### Using dlt (data load tool)

[dlt](https://dlthub.com/) is a Python library that simplifies building data pipelines. You can use dlt to extract data from various sources and load it into formats compatible with Rill.

**Example using dlt:**

```python
import dlt
from dlt.sources.helpers import requests

@dlt.source
def api_source():
    @dlt.resource
    def get_data():
        response = requests.get('https://api.example.com/data')
        yield response.json()
    
    return get_data()

# Run the pipeline and save to Parquet
pipeline = dlt.pipeline(
    pipeline_name='my_pipeline',
    destination='filesystem',
    dataset_name='rill_data'
)

load_info = pipeline.run(api_source())
```

The dlt pipeline will output Parquet files that you can then reference in your Rill models.

### Running Scripts Automatically

You can set up your Python scripts to run automatically using:

- **Cron jobs**: Schedule scripts to run at regular intervals
- **Task schedulers**: Use tools like Apache Airflow, Prefect, or Dagster
- **GitHub Actions**: Run scripts as part of CI/CD pipelines
- **Rill refresh schedules**: Configure model refresh schedules in Rill to trigger your scripts

**Example with Rill model refresh:**

```yaml
type: model
materialize: true

connector: duckdb

sql: |
  select * from read_parquet('data/pipeline_output/*.parquet')

refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

## Best Practices

- **Output formats**: Prefer Parquet for large datasets, CSV for smaller datasets, and JSON for nested data structures
- **Error handling**: Implement robust error handling and logging in your Python scripts
- **Incremental loads**: Design scripts to support incremental data extraction when possible
- **Data validation**: Validate data quality before writing output files
- **Environment variables**: Use environment variables for API keys and credentials
- **Script location**: Keep Python scripts in a `scripts/` directory in your Rill project

## Common Use Cases

- **API data extraction**: Pull data from REST APIs, GraphQL endpoints, or webhooks
- **Database exports**: Extract data from databases that don't have direct Rill connectors
- **File processing**: Process and transform files before ingestion
- **Data enrichment**: Combine multiple data sources before loading into Rill
- **Custom integrations**: Connect to proprietary systems or legacy applications

## Reference

For more information on Python data pipeline tools:
- [dlt documentation](https://dlthub.com/docs)
- [DuckDB Python API](https://duckdb.org/docs/api/python/overview)
- [Rill Models documentation](/build/models)
