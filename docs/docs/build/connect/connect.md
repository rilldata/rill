---
title: "Connect Sources"
description: Import local files or remote data sources
sidebar_label: "Connect Sources"
sidebar_position: 00
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

:::note Full List of Connectors
Rill is continually adding new sources and defined connectors. For a full list, visit our [Connectors](/reference/connectors) page
:::


Rill supports several connectors for importing data: local files, download from an S3 or GCS bucket, download using HTTP(S), connect to databases like MotherDuck or BigQuery. Rill can ingest `.csv`, `.tsv`, `.json`,and `.parquet` files, which may be compressed (`.gz`). 

You can also push logic into source definition during important to filter the data for your source (see Using Code below).



:::tip Import from multiple files
To import data from multiple files, you can use a glob pattern to specify the files you want to include. To learn more about the syntax and details of glob patterns, please refer to the documentation on [glob patterns](/reference/glob-patterns).
:::

## Adding a local file

### Using the UI

To import a file using the UI, click "+" by Sources in the left hand navigation pane, select "Local File", and navigate to the specific file. Alternately, try dragging and dropping the file directly onto the Rill interface.

### Using code
When you add a source using the UI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory. However, you can also create sources more directly by creating the artifact.

In your Rill project directory, create a `source_name.yaml` file in the `sources` directory with the following content:

```yaml
type: local_file
path: /path/to/data.csv
```

Rill will ingest the data next time you run `rill start`.

Note that if you provide a relative path, _the path should be relative to your Rill project root_ (where your `rill.yaml` file is located), **not** relative to the `sources` directory.

:::tip Source Properties

For more details about available configurations and properties, check our [Source YAML](../reference/project-files/sources) reference page.

:::

## Adding a remote source

### Using the UI
To add a remote source using the UI, click "+" by Sources in the left hand navigation pane and select the location where your remote files are stored ("Google Cloud Storage", "Amazon S3", or "http(s)"). Enter your file's URI and click "Add Source".

After import, you can reimport your data whenever you want by clicking the "refresh source" button in the Rill UI.

### Using code
When you add a source using the UI or CLI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory.

To create a remote http(s) source, create a `source_name.yaml` file in the `sources` directory with the following contents:

```yaml
type: https
uri: https://data.example.org/path/to/file.parquet
```
You can also push filters to your source definition using inline editing. Common use cases for inline editing:

- Filter data to only use part of source files
- Push transformations for key fields to source (particularly casting time fields, data types)
- Resolve ingestion issues by declaring types (examples: STRUCT with different values to VARCHAR, fields mixed with INT and VARCHAR values)

:::tip Source Properties

For more details about available configurations and properties, check our [Source YAML](../reference/project-files/sources) reference page.

:::

## Authenticating remote sources

Rill requires credentials to connect to remote data sources such as private buckets in S3, GCS or Azure.

When running Rill locally, Rill attempts to find existing credentials configured on your computer. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

For more details on configuring sources, continue to [Source Credentials](../deploy/credentials).



<!-- WARNING: There are links to this heading in source code. If you change it, find and replace the links. -->

