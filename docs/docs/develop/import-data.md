---
title: Import data source
description: Import local files or remote data sources
sidebar_label: Import data source
sidebar_position: 10
---

Rill supports several connectors for importing data: local files, download from an S3 or GCS bucket, or download using HTTP(S). Rill can ingest `.csv`, `.tsv`, and `.parquet` files, which may be compressed (`.gz`). You can only import a single data file as a source at a time.

## Adding a local file

### Using the UI

To import a file using the UI, click "+" by Sources in the left hand navigation pane, select "Local File", and navigate to the specific file. Alternately, try dragging and dropping the file directly onto the Rill interface.

*Experimental: Alternatively, you can directly query sources from within the [model](./sql-models) itself using a `FROM` statment and `path` with double quotes around it.*

```
FROM "/path/to/data.csv"
```

### Using the CLI

You can also add a local file directly using the Rill CLI. To do so, `cd` into your Rill project and run:
```
rill source add /path/to/file.csv
```

We recommend only using the CLI to import data when the Rill web app is *not* running. 

### Using code
When you add a source using the UI or CLI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory. However, you can also create sources more directly by creating the artifact.

In your Rill project directory, create a `source_name.yaml` file in the `sources` directory with the following contents:

```yaml
type: local_file
path: /path/to/data.csv
```

Rill will ingest the data next time you run `rill start`.

Note that if you provide a relative path, the path should be relative to your Rill project root (where your `rill.yaml` file is located), not relative to the `sources` directory.

## Adding a remote source

### Using the UI
To add a remote source using the UI, click "+" by Sources in the left hand navigation pane and select the location where your remote files are stored ("Google Cloud Storage", "Amazon S3", or "http(s)"). Enter your file's URI and click "Add Source".

*Experimental: Alternatively, you can directly query sources from within the [model](./sql-models) itself using a `FROM` statment and `uri` with double quotes around it. If you need to parameterize your URI for region authentication, we recommend using the modal.*

```
FROM "https://data.example.org/path/to/file.parquet"
```

After import, you can reimport your data whenever you want by clicking the "refresh source" button in the Rill UI.

### Using the CLI
Creating remote sources is not currently available through the CLI.

### Using code
When you add a source using the UI or CLI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory.

To create a remote http(s) source, create a `source_name.yaml` file in the `sources` directory with the following contents:

```yaml
type: https
uri: https://data.example.org/path/to/file.parquet
```

For details about all available properties for all remote connectors, see the syntax [reference](../reference/project-files/sources).

## Authenticating remote sources

Rill requires credentials to connect to remote data sources such as private buckets in S3 or GCS.

When running Rill locally, Rill attempts to find existing credentials configured on your computer. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

Please consult the relevant documentation for instructions on how to configure credentials:

- [Amazon S3](../connectors/s3.md)
- [Google Cloud Storage (GCS)](../connectors/gcs.md)
