---
title: Expressions
description: Details about Rill's expression engine and syntax
sidebar_label: Expressions
sidebar_position: 13
---

Some configurations in Rill can be enabled conditionally through an `if:` field in the related YAML clause. The value provided to `if:` must be a boolean expression evaluating to `true` or `false`. Rill uses [govaluate](https://github.com/Knetic/govaluate) as the engine to parse and evaluate these expressions.

## Examples

- Always: `true`
- Never: `false`
- Using a templating variable: `{{ .env.dev }} == 1`
- Multiple conditions: `'{{ .user.domain }}' == 'example.com' || '{{ .user.domain }}' == 'example.org'`

## Resources

- [govaluate MANUAL.md](https://github.com/Knetic/govaluate/blob/master/MANUAL.md)
