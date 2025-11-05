---
title: Sample Library
sidebar_position: 30
---

# Sample Library

Welcome to the Rill Sample Library! This curated collection of production-ready code examples provides complete, working configurations and snippets that you can copy directly into your Rill projects.

## What's in the Sample Library?

Each sample in this library includes:
- ✅ **Complete, working code** that you can copy and use
- 🏷️ **Tags** to help you find related examples
- 📖 **Documentation links** to learn more about each feature
- 🔒 **Content hashing** to ensure integrity

## Available Samples

### 🔌 Connectors

Complete connector configurations for popular data sources:

- [**S3 Connector**](/examples/sample_library/connector/connector_s3) - Amazon S3 with access keys
- [**Google Cloud Storage**](/examples/sample_library/connector/connector_gcs_service_account) - GCS with service account
- [**Google Cloud Storage (HMAC)**](/examples/sample_library/connector/connector_gcs_hmac) - GCS with HMAC keys
- [**Azure Storage (SAS Token)**](/examples/sample_library/connector/connector_azure_sas_token) - Azure with SAS token
- [**Azure Storage (Account Key)**](/examples/sample_library/connector/connector_azure_storage_account_key) - Azure with account key
- [**BigQuery**](/examples/sample_library/connector/connector_bigquery) - Google BigQuery connector
- [**Athena**](/examples/sample_library/connector/connector_athena) - AWS Athena connector
- [**Google Sheets**](/examples/sample_library/connector/connector_google_sheets) - Google Sheets integration
- [**OpenAI**](/examples/sample_library/connector/connector_openai) - OpenAI API connector
- [**External DuckDB**](/examples/sample_library/connector/connector_external_duckdb) - External DuckDB database

### 📊 Metrics Views

Advanced metrics view configurations:

- [**Complete Metrics View Example**](/examples/sample_library/metrics/example_metrics) - Full metrics view with measures and dimensions
- [**Rolling Averages**](/examples/sample_library/metrics/rolling_average) - Time-window calculations
- [**Metrics Caching**](/examples/sample_library/metrics/metrics_cache) - Performance optimization with caching
- [**Uplift Analysis**](/examples/sample_library/metrics/uplift_using_required_fields) - Using required fields for uplift calculations

### 🔍 Explore Dashboards

Dashboard configuration examples:

- [**Minimal Explore Dashboard**](/examples/sample_library/explore/minimal_explore) - Simplest dashboard setup
- [**Explore with Defaults**](/examples/sample_library/explore/explore_with_defaults) - Dashboard with default configurations

### 🗃️ Data Models

SQL and YAML modeling examples:

**ClickHouse**
- [**ClickHouse Modeling Example**](/examples/sample_library/model/clickhouse/clickhouse_modelling_example) - Complete ClickHouse model
- [**ClickHouse Code Snippets**](/examples/sample_library/model/clickhouse/clickhouse_modelling_snippets) - Useful ClickHouse patterns

### 🔐 Security

Access control and security configurations:

- [**General Access Controls**](/examples/sample_library/security/general_access_controls) - Project-level security
- [**Row-Level Filters**](/examples/sample_library/security/row_level_filters) - Dynamic row filtering based on user attributes
- [**Column-Level Filters**](/examples/sample_library/security/column_level_filters) - Control column visibility

### ⚙️ Project Configuration

Project-level settings and environment management:

- [**Example rill.yaml**](/examples/sample_library/project/example_rill_yaml) - Complete project configuration file
- [**Environment Variables**](/examples/sample_library/project/example_env) - Managing secrets and configuration with .env

### 🔌 APIs

Custom API integration examples:

- [**Metrics SQL API**](/examples/sample_library/api/metrics_sql_api) - Query your metrics views with SQL

## How to Use These Samples

### 1. **Copy and Paste**
Most samples are complete files that you can copy directly into your Rill project.

### 2. **Customize for Your Needs**
Replace placeholder values (like `{{ .env.connector.s3.aws_access_key_id }}`) with your actual configuration or environment variables.

### 3. **Follow the Documentation**
Each sample includes a link to the full documentation for deeper understanding.

### 4. **Mix and Match**
Combine multiple samples to build complex configurations.

## Example: Setting up S3 Connector

1. Navigate to [S3 Connector Sample](/examples/sample_library/connector/connector_s3)
2. Copy the YAML configuration
3. Create a file `connectors/s3.yaml` in your project
4. Add your AWS credentials to `.env`:
   ```bash
   connector.s3.aws_access_key_id=<your-key-id>
   connector.s3.aws_secret_access_key=<your-secret-key>
   ```
5. Reference the connector in your source files

## Tags Reference

Samples are tagged for easy discovery:
- `connector` - Data source and OLAP connectors
- `metrics` - Metrics view configurations
- `model` - Data transformation models
- `security` - Access control and filters
- `code` - Code examples
- `complete_file` - Complete, ready-to-use files
- `snippets` - Partial code snippets

## Need More Examples?

- 📖 **Full Documentation**: Visit [docs.rilldata.com](https://docs.rilldata.com)
- 🎓 **Tutorial Examples**: Check out [Example Projects](https://github.com/rilldata/rill-examples)
- 💬 **Community**: Join our [Discord](https://discord.gg/DJ5qcsxE2m) to share and discuss examples
- 🐛 **Report Issues**: Found a problem? [Open an issue](https://github.com/rilldata/rill/issues)

## Contributing Samples

Have a great example to share? We'd love to include it in the sample library! Check our [Contributing Guide](https://github.com/rilldata/rill/blob/main/CONTRIBUTING.md) to learn how.
