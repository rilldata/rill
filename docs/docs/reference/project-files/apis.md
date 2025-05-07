---
title: API YAML
sidebar_label: API YAML
sidebar_position: 50
---

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition.
See comprehensive documentation on how to define and use [custom APIs](/integrate/custom-apis/index.md)

## Properties

_**`type`**_ — Refers to the resource type and must be `api` _(required)_.

_**`connector`**_ — Refers to the OLAP engine if not already set in rill.yaml or if using [multiple OLAP connectors](../olap-engines/multiple-olap.md) in a single project. Only applies when using `sql` _(optional)_.

Either one of the following:

- _**`sql`**_ — General SQL query referring a [model](/build/models/models.md) _(required)_.

- _**`metrics_sql`**_ — SQL query referring metrics definition and dimensions defined in the [metrics view](/build/dashboards/dashboards.md) _(required)_.

