---
title: Dev/Prod Connector Environments
description: Dev/Prod Setup
sidebar_label: Dev/Prod Connector Setup
sidebar_position: 19
---

Rill uses the Go programming language's [native templating engine](https://pkg.go.dev/text/template), known as `text/template`, which you might know from projects such as [Helm](https://helm.sh/) or [Hugo](https://gohugo.io/). It additionally includes the [Sprig](http://masterminds.github.io/sprig/) library of utility functions.

Templating can be a powerful tool to help introduce dynamic conditional statements based on local variables that have been passed in to Rill or based on the environment being used. Some common use cases may include but are not limited to:
- Defining an [**environment specific database / cluster**](/connect/templating#environment-specific-connectors) to connect to between development and production
- Pointing to [**different source data endpoints / databases**](/connect/templating#environment-specific-data-source-location) between your development and production environments


:::info Where can you template in Rill?

For the most part, templating should be used in [SQL models](../build/models/models.md) and when defining [connector properties](/connect). If you have further questions about templating, please don't hesitate to [reach out](/contact) and we'd love to assist you further!

:::


## Why Use Templating?

Templating serves several important purposes in your data workflow:

- **Cost Management**: Avoid running expensive queries on production data during development
- **Environment Separation**: Keep development and production data sources completely separate
- **Testing Safety**: Test your models and transformations without affecting production data

:::note Rill Developer is a dev environment

Unless explicitly defined, Rill Developer will use a `dev` environment. If you want to emulate production locally, you can do so via `rill start --environment prod`.

:::

## Setting up Environmental Variables

You can set up environmental variables in several locations in Rill. Please review our [configure local credentials documentation](/connect/credentials#setting-credentials-for-rill-developer) for more information.

## Environment Specific Connectors

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

Some connectors will reference two unique databases and require two unique credentials for dev and prod. In this case, you can either define an unique environmental variable for both and reference them separately in the connector.

IE: 
```
"{{ .env.connector.dev_clickhouse.username }}"
"{{ .env.connector.prod_clickhouse.username }}"
```
Or, by creating a separate connector altogether.

```yaml
#/connectors/dev_clickhouse.yaml
#/connectors/prod_clickhouse.yaml
```

:::


## Environment Specific Data Source Location

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

## Additional resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
