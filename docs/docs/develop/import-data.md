---
title: Import data source
description: Import local files or remote data sources
sidebar_label: Import data source
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

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

### Configure credentials for GCS

<!-- WARNING: There are links to this heading in source code. If you change it, find and replace the links. -->

Rill uses the credentials configured in your local environment using the Google Cloud CLI (`gcloud`). Follow these steps to configure it:

1. Open a terminal window and run `gcloud auth list` to check if you already have the Google Cloud CLI installed and authenticated. 

2. If it did not print information about your user, follow the steps on [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk). Make sure to run `gcloud init` after installation as described in the tutorial.

You have now configured Google Cloud access from your local environment. Rill will detect and use your credentials next time you try to ingest a source.

### Configure credentials for S3

<!-- WARNING: There are links to this heading in source code. If you change it, find and replace the links. -->

Rill uses the credentials configured in your local environment using the AWS CLI. 

To check if you already have the AWS CLI installed and authenticated, open a terminal window and run:
```bash
aws iam get-user --no-cli-pager
```
If it prints information about your user, there is nothing more to do. Rill will be able to connect to any data in S3 that you have access to.

If you do not have the AWS CLI installed and authenticated, follow these steps:

1. Open a terminal window and [install the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) if it is not already installed on your system.

2. If your organization has SSO configured, reach out to your admin for instructions on how to authenticate using `aws sso login`.

3. If your organization does not have SSO configured:

    a. Follow the steps described in [How to create an AWS service account using the AWS Management Console](../deploy/credentials/s3.md#how-to-create-an-aws-service-account-using-the-aws-management-console) in our tutorial for [Amazon S3 credentials](../deploy/credentials/s3.md).

    b. Run the following command and provide the access key, access secret, and default region when prompted (you can leave the "Default output format" blank):
    ```
    aws configure
    ```

You have now configured AWS access from your local environment. Rill will detect and use your credentials next time you try to ingest a source.
