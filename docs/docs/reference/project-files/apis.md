---
title: API YAML
sidebar_label: API YAML
sidebar_position: 10
---

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition.
See comprehensive documentation on how to define and use [custom APIs](../../develop/custom-apis/index)

## Properties

_**`kind`**_ — should always be `api` _(required)_

Either one of the following:

- _**`sql`**_ — General SQL query referring a [model](../../develop/sql-models.md) _(required)_

- _**`metrics_sql`**_ — SQL query referring metrics definition and dimensions defined in the [metrics view](../../develop/metrics-dashboard.md) _(required)_