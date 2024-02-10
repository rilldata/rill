---
title: Templating
description: Details about Rill's templating engine and syntax
sidebar_label: Templating
sidebar_position: 12
---

Rill uses the Go programming language's native templating engine, known as `text/template`, which you might know from projects such as [Helm](https://helm.sh/) or [Hugo](https://gohugo.io/). It additionally includes the [Sprig](http://masterminds.github.io/sprig/) library of utility functions.

## Example

Access an environment variable provided using `rill start --variable key=value`:
```sql
SELECT * FROM my_source WHERE foo = '{{ .env.key }}'
```

## Useful resources

- [Official docs](https://pkg.go.dev/text/template) (Go)
- [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) (HashiCorp)
- [Sprig Function Documentation](http://masterminds.github.io/sprig/)
