---
title: Development/Production Connector Environments
description: Development and Production Setup
sidebar_label: Development/Production Connectors
sidebar_position: 19
---

Templating serves several important purposes in your data workflow:

- **Cost Management**: Avoid running expensive queries on production data during development
- **Environment Separation**: Keep development and production data sources completely separate
- **Testing Safety**: Test your models and transformations without affecting production data

:::note Rill Developer defaults to 'dev'

Unless explicitly defined, Rill Developer will use a `dev` environment. If you want to emulate production locally, you can do so via `rill start --environment prod`.

:::

## Setting Up Environmental Variables

You can set up environmental variables in several locations in Rill. Please review our [configure local credentials documentation](/developers/build/connectors/credentials#setting-credentials-for-rill-developer) for more information.

## Referencing Environment Variables

Reference environment variables in your YAML files using the `{{ env "VARIABLE_NAME" }}` syntax:

```yaml
password: '{{ env "POSTGRES_PASSWORD" }}'
google_application_credentials: '{{ env "GOOGLE_APPLICATION_CREDENTIALS" }}'
aws_access_key_id: '{{ env "AWS_ACCESS_KEY_ID" }}'
```

:::tip Case-Insensitive Lookups
The `{{ env "VAR_NAME" }}` function provides case-insensitive variable lookups, so `{{ env "my_var" }}` will match `MY_VAR` in your `.env` file.

**Note:** If your `.env` file contains multiple variables that differ only by case (e.g., both `my_var` and `MY_VAR`), the behavior is undefined. Avoid defining variables with the same name in different cases.
:::

## Environment-Specific Connectors

The most common use case for connector templating is defining separate databases for your development and production operations. This approach gives you the freedom to experiment, test, and iterate on your models without the risk of accidentally modifying or corrupting your production data.

### Example: ClickHouse Connector with Environment Separation

Here's how you can configure a ClickHouse connector to use different environments:

```yaml
type: connector
driver: clickhouse

dev:
  dsn: "clickhouse://user:password@localhost:9000/dev_database" # ClickHouse connection DSN  

# Production environment configuration
prod:
  host: '{{ env "CLICKHOUSE_HOST" }}'
  port: '{{ env "CLICKHOUSE_PORT" }}'
  database: '{{ env "CLICKHOUSE_DATABASE" }}'
  username: '{{ env "CLICKHOUSE_USERNAME" }}'
  password: '{{ env "CLICKHOUSE_PASSWORD" }}'
  ssl: true
  cluster: '{{ env "CLICKHOUSE_CLUSTER" }}'
```

In this example:
- **Development**: Uses a managed ClickHouse instance that Rill starts locally
- **Production**: Connects to your existing ClickHouse cluster using environment variables for secure configuration

:::warning Managing development and production credentials

Some connectors will reference two unique databases and require two unique credentials for development and production. In this case, you can either define a unique environmental variable for both and reference them separately in the connector.

For example:
```yaml
'{{ env "DEV_CLICKHOUSE_USERNAME" }}'
'{{ env "PROD_CLICKHOUSE_USERNAME" }}'
```

Or, by creating a separate connector altogether.

```yaml
# /connectors/dev_clickhouse.yaml
# /connectors/prod_clickhouse.yaml
```

:::

## Environment-Specific Data Source Location

Rill allows you to configure different data source locations for different environments (development, staging, production) using templating. This enables you to work with development datasets locally while pointing to production data sources in cloud deployments.

```yaml
type: connector
driver: bigquery
dev:
  project_id: rilldata_dev
project_id: rilldata

google_application_credentials: '{{ env "GOOGLE_APPLICATION_CREDENTIALS" }}'
```

```yaml
type: connector                                  
driver: postgres                                
dev:
  dsn: "postgres://user:password@localhost:5432/dev_database"
dsn: "postgres://user:password@localhost:5432/database"
```

## Additional Resources
- [Performance Optimization Guide](/developers/guides/performance)
- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
