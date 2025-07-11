---
title: Templating and Environments
description: Details about Rill's templating engine and syntax
sidebar_label: Model Templating
sidebar_position: 35
---
## Overview

Rill uses the Go programming language's [native templating engine](https://pkg.go.dev/text/template), known as `text/template`, which you might know from projects such as [Helm](https://helm.sh/) or [Hugo](https://gohugo.io/). It additionally includes the [Sprig](http://masterminds.github.io/sprig/) library of utility functions.


## Environments and Rill

Environments allow the separation of logic between different deployments of Rill, most commonly when using a project in Rill Developer and Rill Cloud. Generally speaking, Rill Developer is meant primarily for local development purposes, where the developer may want to use a segment or sample of data for modeling purposes to help validate the model logic is correct and producing expected results. Then, once finalized and a project is ready to be deployed to Rill Cloud, the focus is on shared collaboration at scale and where most of the dashboard consumption in production will happen.

There could be a few reasons to have separate logic for sources, models, and dashboards for Rill Developer and Rill Cloud, respectively:
1. As the size of data starts growing, the locally embedded OLAP engine (using DuckDB) may start to face scaling challenges, which can impact performance and the "snappiness" of models and dashboards. Furthermore, for model development, the full data is often not needed and working with a sample is sufficient. For production though, where analysts and business users are interacting with Rill dashboards to perform interactive, exploratory analysis or make decisions, it is important that these same models and dashboards are powered by the entire range of data available.
2. For certain organizations, there might be a development and production version of source data. Therefore, you can develop your models and validate the results in Rill Developer against development. When deployed to Rill Cloud, these same models and dashboards can then be powered by your production source, reflecting the most correct and up-to-date version of your business data.


```sql
# models/my_model.sql
SELECT *
FROM {{ ref "my_source" }}
{{ if dev }}
WHERE my_value = '{{ .env.my_value }}'
{{ end }}
```

In this example, the `ref` tag ensures that the model `my_model` will not be created until **after** a source named `my_source` has finished ingesting.

## Using environments to generate custom templated SQL

Environments are also useful when you wish to apply environment-specific SQL logic to your sources and models. One common use case would be to apply a filter or limit for models automatically when developing locally (in Rill Developer), but not have these same conditions applied to production models deployed on Rill Cloud. These same principles could also be extended to apply more advanced logic and conditional statements based on your requirements. This is all possible by combining environments with Rill's ability to leverage [templating](/connect/templating.md).

Let's say we had some kind of intermediate model where we wanted to apply a limit of 10000 _(but only for local development)_, our `model.sql` file may look something like the following instead:

```sql
SELECT * FROM {{ ref "<source_name>" }}
{{ if dev }} LIMIT 10000 {{ end }}
```

:::warning When applying templated logic to model SQL, make sure to leverage the `ref` function

If you use templating in SQL models, you must replace references to tables / models created by other sources or models with `ref` tags. See this section on ["Referencing other tables or models in SQL when using templating"](/connect/templating.md#referencing-other-tables-or-models-in-sql-when-using-templating). This ensures that the native Go templating engine used by Rill is able to resolve and correctly compile the SQL syntax during runtime (to avoid any potential downstream errors).

:::

## Examples

Let's walk through a few example scenarios to illustrate the power of templating and how it can be used within Rill.



### Limiting the number of rows in a model only for local development

Following a similar vein to our previous example, let's say that we wanted to apply a limit (or other custom SQL) to our models that only came into effect when used in development (i.e. Rill Developer), but not in production (i.e. Rill Cloud). A very straightforward example would be that perhaps we have some complex SQL written that is computationally intensive and our source data is quite large. For local modeling purposes, we don't need to work with the full dataset but need to only return the first 1000 rows to validate that the logic is correct and that results are returning as expected.

In our `model.sql` file, we could leverage templating logic to check the environment is `dev` and apply a `LIMIT 1000` to the query:

```sql
-- A bunch of CTEs, complex joins, etc.

SELECT * FROM final
{{if dev}} LIMIT 1000 {{end}}
```

```yaml
dev:
   sql: SELECT * FROM final LIMIT 1000

sql: SELECT * FROM final
```

:::tip Running in Production
Running the same model above in Rill Cloud, the full dataset will be used and this limit will only apply to local development with Rill Developer!
:::

### Leveraging variables to apply a filter and row limit dynamically to a model

Our last example will highlight how the same templating concepts can be applied with [variables](#setting-variables-in-rill) instead of [environments](../transform/models/environments). In this case, we have a source dataset about horror movies that came out in the past 50 years, which includes various characteristics, attributes, and metrics about each horror movie as separate columns. For example, we know the release date, how many people saw a movie, what the budget was, it's popularity, the original language of the movie, the genres, and much more.

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

If you use templating in SQL models, you must replace references to tables / models created by other sources or models with `ref` tags. See this section on ["Referencing other tables or models in SQL when using templating"](#referencing-other-tables-or-models-in-sql-when-using-templating). This ensures that the native Go templating engine used by Rill is able to resolve and correctly compile the SQL syntax during runtime (to avoid any potential downstream errors).

:::

If we simply run Rill Developer using `rill start`, our model will look like the following (this will also reflect our data model in production, i.e. Rill Cloud, after we've [pushed the changes for the project to GitHub](./deploy-dashboard/)):

<img src = '/img/deploy/templating/vars-example.png' class='rounded-gif' />
<br />


**Now**, just to illustrate what a local override might look like, let's say we stop Rill Developer and then restart Rill via the CLI with the following command:
```bash
rill start --env language="ja" --env local_limit=100
```

Even though we have defaults set in `rill.yaml` (and this will be used by any downstream models and dashboards on Rill Cloud), we will instead see these local overrides come into effect with our templated logic to return Spanish movies and the model limit is now 100 rows.

<img src = '/img/deploy/templating/vars-override-example.png' class='rounded-gif' />
<br />


## Additional resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)

