---
title: SQL API
sidebar_label: SQL API
sidebar_position: 40
---

You can write a SQL query and expose it as an API endpoint. This is useful when you want to directly write queries 
against a [model](/build/models/models.md) or [source](../../reference/project-files/sources.md) that you have created. 
It should have the following structure:

```yaml
type: api
sql: SELECT abc FROM my_table
```

where `my_table` is your model or source name.

## SQL Templating

You can use templating to make your SQL query dynamic. We support:
 - Dynamic arguments that can be passed in as query params during api call using `{{ .args.<param-name> }}`
 - User attributes like email, domain and admin if available using `{{ .user.<attr> }}` (see integration docs [here](/integrate/custom-api.md) for when user attributes are available)
 - Conditional statements 
 - Optional parameters paired with conditional statements.

See integration docs [here](/integrate/custom-api.md) to learn how to these are passed in while calling the API.

### Conditional statements

Assume a API endpoint defined as `my-api.yaml`:
```yaml
type: api
sql: |
  SELECT count("measure")
    {{ if ( .user.admin ) }} ,dim  {{ end }} 
    FROM my_table WHERE date = '{{ .args.date }}' 
    {{ if ( .user.admin ) }} group by 2 {{ end }}
```

will expose an API endpoint like `https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/my-api?date=2021-01-01`.
If the user is an admin, the API will return the count of `measure` by `dim` for the given date. If the user is not an admin, the API will return the count of `measure` for the given date.


### Optional parameters

Rill utilizes standard Go templating together with [Sprig](http://masterminds.github.io/sprig/) which adds a number of useful utility functions.  
One of those functions is `hasKey` which in the example below enables optional parameters being passed to the Custom API endpoint. This allows you to build API endpoints that can handle a wider range of parameters and logic, reducing need to duplicate API endpoints.

Assume a API endpoint defined as `my-api.yaml`:
```yaml
SELECT
  device_type,
  AGGREGATE(overall_spend)
FROM bids
{{ if hasKey .args "type" }} WHERE device_type = '{{ .args.type }}' {{ end }} 
GROUP BY device_type
```

HTTP Get `.../runtime/api/my-api` would return `overall_spend` for all `device_type`'s  
HTTP Get `.../runtime/api/my-api?type=Samsung` would return `overall_spend` for `Samsung`



## Useful resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)

