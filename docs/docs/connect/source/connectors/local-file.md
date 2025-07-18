---
title: Local File
description: Connect to your local data
sidebar_label: Local File
sidebar_position: 17
---


## Adding a local file

### Using the UI

To import a file using the UI, click "+" by Sources in the left-hand navigation pane, select "Local File", and navigate to the specific file. Alternatively, try dragging and dropping the file directly onto the Rill interface.

### Using code
When you add a source using the UI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory. However, you can also create sources more directly by creating the artifact.

In your Rill project directory, create a `source_name.yaml` file in the `sources` directory with the following content:

```yaml
type: model
connector: local_file
path: /path/to/local/data.csv
```

Rill will ingest the data next time you run `rill start`.

Note that if you provide a relative path, _the path should be relative to your Rill project root_ (where your `rill.yaml` file is located), **not** relative to the `sources` directory.

:::tip Import from multiple files
To import data from multiple files, you can use a glob pattern to specify the files you want to include. To learn more about the syntax and details of glob patterns, please refer to the duckdb documentation on [reading multiple files](https://duckdb.org/docs/stable/data/multiple_files/overview.html).
:::

:::note Source Properties

For more details about available configurations and properties, check our [Source YAML](/reference/project-files/sources) reference page.

:::
