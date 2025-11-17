---
title: Model Environment Templating
description: Dev/Prod Setup
sidebar_label: Environment Templating
sidebar_position: 40
---

Along with [connector templating](/build/connectors/templating), you can also optimize your development environment for performance and cost-effectiveness at the model layer. This approach is particularly useful when you don't have a separate development database or file system, as it allows you to limit the amount of data processed in source models.

## Why Use Templating?

Templating serves several important purposes in your data workflow:

- **Cost Management**: Avoid running expensive queries on production data during development
- **Performance in dev**: Optimize development environment performance by limiting data volume
- **Testing Safety**: Test your models and transformations without affecting production data

:::note Rill Developer is a dev environment

Unless explicitly defined, Rill Developer will use a `dev` environment. If you want to emulate production locally, you can do so via `rill start --environment prod`.

:::

## Separate Dev and Prod SQL Queries

Another common templating pattern is using the `dev:` partition in your model definitions. This tells Rill to use a different data source or query when running in development mode, typically with a smaller dataset or different data location.

### Example: Ingest Data from Specific Folder Path in Dev

```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
connector: "duckdb"

dev:
  sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_06.parquet')"

sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet')"
```

:::info Why is the connector type duckdb and not s3 or gcs?

In this case, we are using the [embedded DuckDB engine](/build/connectors/olap/duckdb) to execute a [SELECT](https://duckdb.org/docs/sql/statements/select.html) statement while leveraging DuckDB's native [read_parquet](https://duckdb.org/docs/data/parquet/overview.html) function. Therefore, the `connector` type ends up being `duckdb` instead of `s3`. For more details, see our [Source YAML](/reference/project-files/sources) reference documentation.

:::

<img src='/img/deploy/templating/gcs-env-example.png' class='rounded-gif' />
<br />

In this example:
- **Development**: Reads from a single file (`2025/06/commits_2025_06.parquet`) for faster processing
- **Production**: Reads from all files in the folder using the `/*/*/` wildcard pattern; you can also use `/**/` if you're unsure of the number of directories

## Inline SQL Templating

You can also use Go template syntax directly within your SQL queries to conditionally include or exclude parts of your query based on the environment.

### Example: Conditional SQL Limiting Dev Rows

```sql
select   
    *
from read_csv('gs://your-bucket/folder/**/file.csv', auto_detect=true, ignore_errors=1, header=true)
{{ if dev }} 
    limit 10000 
{{ end }}
```

```sql
select   
    *
from read_csv('gs://your-bucket/folder/**/file.csv', auto_detect=true, ignore_errors=1, header=true)
{{ if dev }} 
    where column = 'value'
{{ end }}
```

In this example:
- **Development**: Adds a where/limit clause, filtering the data by a specific value to restrict the dataset size
- **Production**: No limits

### Example: If/Else Conditions

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
- **Production**: Applies a where clause to filter the data in production

### Applying a One-Week Sample to the Source Bucket for Local Development

In another example, let's say we had an [S3](/build/connectors/data-source/s3) source defined that happens to be reading a very large amount of parquet data. Following best practices, we'll want to read in a subset of this source data for local modeling in Rill Developer rather than using the full dataset for development purposes. Furthermore, we'll make the assumption that the upstream data is not partitioned and thus the S3 bucket is not partitioned (where we could then simply filter the `path` by using a glob pattern potentially in conjunction with [environment-specific logic](/build/models/templating)). So what can we do?

Fortunately, we can leverage DuckDB's ability to read from S3 files directly and _apply a filter post-download_ using templating logic in the SQL. In this case, because there is an existing `updated_at` timestamp column, we can use it to filter and retrieve only one week's worth of data. For example, our `source.yaml` file may end up looking something like this:

```yaml
type: model
connector: "duckdb"
sql: > 
    SELECT * FROM read_parquet('s3://bucket/path/*.parquet')
    {{ if dev }}
      where updated_at > CURRENT_DATE - INTERVAL 7 DAY
    {{ end }}
```

:::tip Dynamic dates

Depending on the OLAP engine of your project, you can set the dates dynamically using `CURRENT_DATE` (DuckDB) or `now()` (ClickHouse) and subtract days.

:::

### Example: Leveraging Variables to Apply a Filter and Row Limit Dynamically to a Model

Our last example will highlight how the same templating concepts can be applied with [variables](/build/connectors/credentials#setting-credentials-for-rill-developer). In this case, we have a source dataset about horror movies that came out in the past 50 years, which includes various characteristics, attributes, and metrics about each horror movie as separate columns.

For example, we know the release date, how many people saw a movie, what the budget was, its popularity, the original language of the movie, the genres, and much more.

Let's say that we wanted to apply a filter on the resulting model based on the `original_language` of the movie and also limit the number of records that we retrieve, which will be based on the `language` and `local_limit` variables we have defined. Taking a quick look at our project's `rill.yaml` file, we can see the following configuration (to return only English movies and apply a limit of 5):

```yaml
env:
  local_limit: 5
  language: "en"
```

Furthermore, our `model.sql` file contains the following SQL:

```sql
SELECT * FROM {{ ref "snowflake" }}
WHERE original_language = '{{ .env.language }}'
{{if dev}} LIMIT {{ .env.local_limit }} {{end}}
```

:::warning When applying templated logic to model SQL, make sure to leverage the `ref` function

If you use templating in SQL models, you must replace references to tables/models created by other sources or models with `ref` tags. This ensures that the native Go templating engine used by Rill is able to resolve and correctly compile the SQL syntax during runtime (to avoid any potential downstream errors).

:::

If we simply run Rill Developer using `rill start`, our model will look like the following (this will also reflect our data model in production, i.e., Rill Cloud, after we've [pushed the changes for the project to GitHub](/deploy/deploy-dashboard)):

<img src='/img/deploy/templating/vars-example.png' class='rounded-gif' />
<br />

**Now**, just to illustrate what a local override might look like, let's say we stop Rill Developer and then restart Rill via the CLI with the following command:

```bash
rill start --env language="ja" --env local_limit=100
```

Even though we have defaults set in `rill.yaml` (and this will be used by any downstream models and dashboards on Rill Cloud), we will instead see these local overrides come into effect with our templated logic to return Japanese movies and the model limit is now 100 rows.

<img src='/img/deploy/templating/vars-override-example.png' class='rounded-gif' />
<br />
