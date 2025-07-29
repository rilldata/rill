---
title: Dev/Prod Environments
description: Dev/Prod Setup
sidebar_label: Dev/Prod Setup
sidebar_position: 19
---

Templating is a powerful feature in Rill that allows you to seamlessly switch between development and production environments without worrying about ingesting large amounts of data locally. This is particularly useful when working with data warehouses that charge based on query volume, as it helps you avoid unnecessary costs during development and testing phases.

## Why Use Templating?

Templating serves several important purposes in your data workflow:

- **Cost Management**: Avoid running expensive queries on production data during development
- **Environment Separation**: Keep development and production data sources completely separate
- **Data Shaping**: Transform data locally to make it easier to work with during development
- **Testing Safety**: Test your models and transformations without affecting production data

## Different Database Environments

The most common use case for connector templating is defining separate database clusters for your development and production operations. This approach gives you the freedom to experiment, test, and iterate on your models without the risk of accidentally modifying or corrupting your production data.

### Example: ClickHouse Connector with Environment Separation

Here's how you can configure a ClickHouse connector to use different environments:

```yaml
type: connector
driver: clickhouse

# Development environment configuration
dev:
  managed: true  # Rill will start ClickHouse as a subprocess locally

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

## Development vs Production Data Sources

Another common templating pattern is using the `dev:` partition in your source definitions. This tells Rill to use a different data source or query when running in development mode, typically with a smaller dataset or different data location.

### Example: Limited Data in Source in Development

```yaml
type: source

connector: "duckdb"

# Development configuration - uses a single file for faster processing
dev:
    sql: |
        select   
            *
        from read_csv('gs://your-bucket/folder/20250701/file.csv', auto_detect=true, ignore_errors=1, header=true)

# Production configuration - processes all files in the folder
sql: |
    select   
        *
    from read_csv('gs://your-bucket/folder/**/file.csv', auto_detect=true, ignore_errors=1, header=true)
```

In this example:
- **Development**: Reads from a single file (`20250701/file.csv`) for faster processing
- **Production**: Reads from all files in the folder using the `**` wildcard pattern

## Inline SQL Templating

You can also use Go template syntax directly within your SQL queries to conditionally include or exclude parts of your query based on the environment.

### Example: Conditional SQL with Go Templates

```sql
select   
    *
from read_csv('gs://your-bucket/folder/**/file.csv', auto_detect=true, ignore_errors=1, header=true)
{{ if dev }} 
    limit 10000 
{{ else }} 
    where column = 'value' 
{{ end }}
```

In this example:
- **Development**: Adds a `LIMIT 10000` clause to restrict the dataset size
- **Production**: Applies a `WHERE` filter instead of limiting rows

## Best Practices

When using templating in your Rill projects:

1. **Keep Development Data Small**: Use limited datasets in development to speed up iteration
2. **Use Environment Variables**: Store sensitive configuration in environment variables
3. **Test Both Environments**: Ensure your templates work correctly in both dev and prod
4. **Document Your Templates**: Add comments to explain the purpose of different configurations
5. **Version Control**: Include template files in version control but exclude environment-specific files

## Switching Between Environments

By default, Rill Developer runs in development mode. You can manually switch to production mode by passing the `--prod` flag when starting Rill Developer:

```bash
rill start --prod
```

This allows you to test your production configuration locally when needed.