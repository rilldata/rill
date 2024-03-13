---
title: Templating
description: Details about Rill's templating engine and syntax
sidebar_label: Templating
sidebar_position: 12
---

Rill uses the Go programming language's native templating engine, known as `text/template`, which you might know from projects such as [Helm](https://helm.sh/) or [Hugo](https://gohugo.io/). It additionally includes the [Sprig](http://masterminds.github.io/sprig/) library of utility functions.

:::warning
If you use templating in SQL models, you must replace references to tables created by other sources or models with `ref` tags. See the "Referencing other tables in SQL when using templating" section below for details.
:::


## Example

Access an environment variable provided using `rill start --var key=value`:
```sql
SELECT * FROM my_source WHERE foo = '{{ .vars.key }}'
```

## Referencing other tables in SQL when using templating

When you use templating in a SQL model, Rill loses the ability to analyze the SQL for references to other sources and models in the project. This can lead to reconcile errors where Rill tries to create a model before the sources (or other models) it depends upon have finished ingesting.

To avoid this, whenever you start using templating in a model's SQL, it is recommended that you use `ref` tags every time you reference another resource in your project in SQL. For example:
```sql
# models/my_model.sql
SELECT *
FROM {{ ref "my_source" }}
WHERE my_value = '{{ .vars.my_value }}'
```
In this example, the `ref` tag ensures that the model `my_model` will not be created until *after* a source named `my_source` has finished ingesting.

## Useful resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
