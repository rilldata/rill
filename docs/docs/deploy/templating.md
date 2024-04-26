---
title: Templating
description: Details about Rill's templating engine and syntax
sidebar_label: Templating
sidebar_position: 12
---
## Overview

Rill uses the Go programming language's [native templating engine](https://pkg.go.dev/text/template), known as `text/template`, which you might know from projects such as [Helm](https://helm.sh/) or [Hugo](https://gohugo.io/). It additionally includes the [Sprig](http://masterminds.github.io/sprig/) library of utility functions.

Templating can be a powerful tool to help introduce dynamic conditional statements based on local variables that have been passed in to Rill or based on the environment being used. Some common use cases may include but are not limited to:
- Pointing to different source data endpoints / databases between your development and production environments, i.e. Rill Developer vs Rill Cloud
- Working with a sample or subset of data during local development (but making sure the full dataset is being used in production dashboards)
- Applying filters or other if/else predefined logic to run different SQL whether a model is being run locally or in production
- There are many more possibilities!

:::info Where can you template in Rill?

For the most part, templating should be used in [SQL models](../build/models/models.md) and when defining [source proprties](/reference/project-files/sources.md). If you have further questions about templating, please don't hesitate to [reach out](../contact.md) and we'd love to assist you further!

:::

### Setting variables in Rill

Templating can be used in conjunction with variables to apply more advanced logic to your sources and models. 

Variables can be set in Rill through one of the following methods:
1. Defining the corresponding key-value pair under the top-level `vars` key in `rill.yaml`
2. Manually passing in the variable when starting Rill (i.e. `rill start --var <var_name>=<value>`)
3. Specifying the key-value pair for your variable in your `<RILL_PROJECT_HOME>/.env` file and/or using `rill env set` to set the variable via the CLI (and then [pushing / pulling as appropriate](../build/credentials/credentials.md#variables))

For more information, please check our [reference documentation](/reference/project-files/rill-yaml.md#setting-variables).

### Environments and Rill

Templating can be used in conjunction with environments to apply environment-specific logic based on whether the project is running locally on Rill Developer or when deployed to Rill Cloud. By default, Rill comes with two built-in environments defined, `dev` and `prod`, which correspond to Rill Developer and Rill Cloud respectively. For more details about using environments in Rill Cloud, please see our [environments](../build/models/environments.md) page.

### Referencing other tables or models in SQL when using templating

When you use templating in a SQL model, Rill loses the ability to analyze the SQL for references to other sources and models in the project. This can lead to reconcile errors where Rill tries to create a model before the sources (or other models) it depends upon have finished being ingested.

To avoid this scenario, whenever you use templating in a model's SQL, it is <u>strongly recommended</u> to incorporate `ref` tags whenever you need to reference another resource in your project in SQL. For those familiar with [dbt's ref() function](https://docs.getdbt.com/reference/dbt-jinja-functions/ref), the concept is very similar in nature. As an example:

```sql
# models/my_model.sql
SELECT *
FROM {{ ref "my_source" }}
WHERE my_value = '{{ .vars.my_value }}'
```

In this example, the `ref` tag ensures that the model `my_model` will not be created until **after** a source named `my_source` has finished ingesting.

## Examples

Let's walk through a few example scenarios to illustrate the power of templating and how it can be used within Rill.

### Changing the database user based on dev / prod

Let's say that we have a [Snowflake](/reference/connectors/snowflake.md) source created that is being used in some downstream models and dashboards. In this case, we want the following logic applied:
- In local **development**, we want _Rill Developer_ to use our dev user (e.g. `SUPPORT_TEST`) to connect to Snowflake
- In local **development**, we want to limit the size of our source data in _Rill Developer_ (in this case just a simple `LIMIT 1` to illustrate the point)
- In **production**, we want _Rill Cloud_ to use our provisioned service account (e.g. `PROD_USER`) to connect to Snowflake
- In **production**, we want to make sure that _Rill Cloud_ is using the entire source data for our downstream models and dashboards

In this hypothetical scenario, our `source.yaml` might look something like the following:
```yaml
kind: source
connector: "snowflake"
sql: "select * from <table_name> {{if dev}} limit 1 {{end}}"
dsn: "{{if dev}}SUPPORT_TEST{{else}}PROD_USER{{end}}@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>N&authenticator=SNOWFLAKE_JWT&privateKey=..."
```

![Dynamically changing the Snowflake user based on environment](/img/deploy/templating/snowflake-env-example.png)

### Changing the bucket location based on dev / prod

Let's say that we have a [GCS](/reference/connectors/gcs.md) source created where Rill is reading in some CSV data (in this case we have some sample [Citi Bike trip data](https://citibikenyc.com/system-data) loaded onto both a "test" and "prod" GCS bucket). In this case, let's imagine that we want to connect to this "test" bucket for local development purposes but we want to make sure that our production data hosted on our "prod" bucket is what's being used to power this same source once the project has been deployed to Rill Cloud. In such a scenario, our `source.yaml` might look like:

```yaml
kind: source
connector: "duckdb"
sql: "select * from read_csv('gs://{{if dev}}<test_bucket>{{else}}<prod_bucket>{{end}}/201306-citibike-tripdata.csv', auto_detect=true, ignore_errors=1, header=true)"
```

![Dynamically changing the bucket used based on environment](/img/deploy/templating/gcs-env-example.png)


### Applying a one week sample to the source bucket for local development

In another example, let's say we had a [S3](/reference/connectors/s3.md) source defined that happens to be reading a very large amount of parquet data. Following [best practices](performance.md#work-with-a-subset-of-your-source-data-for-local-development-and-modeling), we'll want to read in a subset of this source data for local modeling in Rill Developer rather than using the full dataset for development purposes. Furthermore, we'll make the assumption that the upstream data is not partitioned and thus the S3 bucket is not partitioned (where we could then simply filter the `path` by using a [glob pattern](../build/connect/glob-patterns.md) potentially in conjunction with [environment specific logic](../build/models/environments.md)). So what can we do?

Fortunately, we can leverage DuckDB's ability to read from S3 files directly and _apply a filter post-download_ using templating logic in the SQL. In this case, because there is an existing `updated_at` timestamp column, we can use it to filter and retrieve only one week's worth of data. For example, our `source.yaml` file may end up looking something like:

```yaml
kind: source
connector: "duckdb"
sql: SELECT * FROM read_parquet('s3://bucket/path/*.parquet') {{ if dev }} where updated_at >= '2024-03-01' AND updated_at < '2024-03-07' {{ end }}
```

:::info Why is the connector type duckdb and not s3?

In this case, we are using the [embedded DuckDB engine](/reference/olap-engines/duckdb.md) to execute a [SELECT](https://duckdb.org/docs/sql/statements/select.html) statement while leveraging DuckDB's native [read_parquet](https://duckdb.org/docs/data/parquet/overview.html) function. Therefore, the `connector` type ends up being `duckdb` instead of `s3`. For more details, see our [Source YAML](/reference/project-files/sources.md) reference documentation.

:::

### Limiting the number of rows in a model only for local development

Following a similar vein to our previous example, let's say that we wanted to apply a limit (or other custom SQL) to our models that only came into effect when used in development (i.e. Rill Developer), but not in production (i.e. Rill Cloud). A very straightforward example would be that perhaps we have some complex SQL written that is computationally intensive and our source data is quite large. For local modeling purposes, we don't need to work with the full dataset but need to only return the first 1000 rows to validate that the logic is correct and that results are returning as expected.

In our `model.sql` file, we could leverage templating logic to check the environment is `dev` and apply a `LIMIT 1000` to the query:

```sql
-- A bunch of CTEs, complex joins, etc.

SELECT * FROM final
{{if dev}} LIMIT 1000 {{end}}
```

:::tip Running in Production
Running the same model above in Rill Cloud, the full dataset will be used and this limit will only apply to local development with Rill Developer!
:::

### Leveraging variables to apply a filter and row limit dynamically to a model

Our last example will highlight how the same templating concepts can be applied with [variables](#setting-variables-in-rill) instead of [environments](../build/models/environments.md). In this case, we have a source dataset about horror movies that came out in the past 50 years, which includes various characteristics, attributes, and metrics about each horror movie as separate columns. For example, we know the release date, how many people saw a movie, what the budget was, it's popularity, the original language of the movie, the genres, and much more.

Let's say that we wanted to apply a filter on the resulting model based on the `original_language` of the movie and also limit the number of records that we retrieve, which will be based on the `language` and `local_limit` variables we have defined. Taking a quick look at our project's `rill.yaml` file, we can see the following configuration (to return only English movies and apply a limit of 5):

```yaml
vars:
  local_limit: 5
  language: "en"
```

Furthermore, our `model.sql` file contains the following SQL:

```sql
SELECT * FROM {{ ref "data_source" }}
WHERE original_language = '{{ .vars.language }}'
{{if dev}} LIMIT {{ .vars.local_limit }} {{end}}
```

:::warning When applying templated logic to model SQL, make sure to leverage the `ref` function

If you use templating in SQL models, you must replace references to tables / models created by other sources or models with `ref` tags. See this section on ["Referencing other tables or models in SQL when using templating"](#referencing-other-tables-or-models-in-sql-when-using-templating). This ensures that the native Go templating engine used by Rill is able to resolve and correctly compile the SQL syntax during runtime (to avoid any potential downstream errors).

:::

If we simply run Rill Developer using `rill start`, our model will look like the following (this will also reflect our data model in production, i.e. Rill Cloud, after we've [pushed the changes for the project to Github](./existing-project/existing-project.md)):

![Using templating logic with variables to create custom SQL](/img/deploy/templating/vars-example.png)

**Now**, just to illustrate what a local override might look like, let's say we stop Rill Developer and then restart Rill via the CLI with the following command:
```bash
rill start --var language="es" --var local_limit=100
```

Even though we have defaults set in `rill.yaml` (and this will be used by any downstream models and dashboards on Rill Cloud), we will instead see these local overrides come into effect with our templated logic to return Spanish movies and the model limit is now 100 rows.

![Using templating logic with local variable overrides for our model SQL](/img/deploy/templating/vars-override-example.png)

Voila!

## Additional resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
