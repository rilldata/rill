---
title: Dev/Prod Connector Environments
description: Dev/Prod Setup
sidebar_label: Dev/Prod Connectors
sidebar_position: 19
---

Templating serves several important purposes in your data workflow:

- **Cost Management**: Avoid running expensive queries on production data during development
- **Environment Separation**: Keep development and production data sources completely separate
- **Testing Safety**: Test your models and transformations without affecting production data

:::note Rill Developer is a dev environment

Unless explicitly defined, Rill Developer will use a `dev` environment. If you want to emulate production locally, you can do so via `rill start --environment prod`.

:::

## Setting Up Environmental Variables

You can set up environmental variables in several locations in Rill. Please review our [configure local credentials documentation](/build/connectors/credentials#setting-credentials-for-rill-developer) for more information.

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
  host: "{{ .env.connector.clickhouse.host }}"
  port: "{{ .env.connector.clickhouse.port }}"
  database: "{{ .env.connector.clickhouse.database }}"
  username: "{{ .env.connector.clickhouse.username }}"
  password: "{{ .env.connector.clickhouse.password }}"
  ssl: true
  cluster: "{{ .env.connector.clickhouse.cluster }}"
```

In this example:
- **Development**: Uses a managed ClickHouse instance that Rill starts locally
- **Production**: Connects to your existing ClickHouse cluster using environment variables for secure configuration

:::warning Managing dev and prod credentials

Some connectors will reference two unique databases and require two unique credentials for dev and prod. In this case, you can either define a unique environmental variable for both and reference them separately in the connector.

IE:
```
"{{ .env.connector.dev_clickhouse.username }}"
"{{ .env.connector.prod_clickhouse.username }}"
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

google_application_credentials: '{{ .env.connector.bigquery.google_application_credentials}}'
```

```yaml
type: connector                                  
driver: postgres                                
dev:
  dsn: "postgres://user:password@localhost:5432/dev_database"
dsn: "postgres://user:password@localhost:5432/database"
```

## Additional Resources
- [Performance Optimization Guide](/guides/performance)
- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
