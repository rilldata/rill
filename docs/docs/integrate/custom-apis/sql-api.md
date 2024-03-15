---
title: SQL API
sidebar_label: SQL API
sidebar_position: 40
---

You can write a SQL query and expose it as an API endpoint. This is useful when you want to directly write queries 
against a [model](/build/models/models.md) or [source](../../reference/project-files/sources.md) that you have created. 
It should have the following structure:

```yaml
kind: api
sql: SELECT abc FROM my_table
```

where `my_table` is your model or source name.

## SQL Templating

You can use templating to make your SQL query dynamic. We support:
 - Dynamic arguments that can be passed in as query params during api call using `{{ .args.<param-name> }}`
 - User attributes like email, domain and admin if available using `{{ .user.<attr> }}` (see integration docs [here](/integrate/custom-api.md) for when user attributes are available)
 - Conditional statements using go ang sprig templating functions (see resources at the end for more details) 

For example:

`my-api.yaml`:
```yaml
kind: api
sql: |
  SELECT count("measure")
    {{ if ( .user.admin ) }} ,dim  {{ end }} 
    FROM my_table WHERE date = '{{ .args.date }}' 
    {{ if ( .user.admin ) }} group by 2 {{ end }}
```

will expose an API endpoint like `https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/my-api?date=2021-01-01`.
If the user is an admin, the API will return the count of `measure` by `dim` for the given date. If the user is not an admin, the API will return the count of `measure` for the given date.

See integration docs [here](/integrate/custom-api.md) to learn how to these are passed in while calling the API.

## Useful resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)

