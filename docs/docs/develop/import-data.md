---
title: "Import: Add Sources"
description: Import local files or remote data sources
sidebar_label: "Import: Add Sources"
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

Rill supports several connectors for importing data: local files, download from an S3 or GCS bucket, download using HTTP(S), connect to databases like MotherDuck or BigQuery. Rill can ingest `.csv`, `.tsv`, `.json`,and `.parquet` files, which may be compressed (`.gz`). 

:::tip Import from multiple files
To import data from multiple files, you can use a glob pattern to specify the files you want to include. To learn more about the syntax and details of glob patterns, please refer to the documentation on [glob patterns](/reference/glob-patterns).
:::

You can also push logic into source definition during important to filter the data for your source (see Using Code below).

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

### Google Sheets

Rill is very flexible and can read from any http(s) URL endpoint that produces a valid data file in a supported format. For example, to bring in data from Google Sheets as a CSV file directly into Rill as a source ([leveraging the direct download link syntax](https://www.highviewapps.com/blog/how-to-create-a-csv-or-excel-direct-download-link-in-google-sheets/)), you can create a `source_name.yaml` file in the `sources` directory of your Rill project directory with the following content:

```yaml
type: "duckdb"
path: "select * from read_csv_auto('https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/export?format=csv&gid=<SHEET_ID>', normalize_names=True)"
```

:::note Updating the URL

Make sure to replace `SPREADSHEET_ID` and `SHEET_ID` with the ID of your spreadsheet and tab respectively (which you can obtain from looking at the URL when Google Sheets is open).

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

### Configure credentials for GCS

<!-- WARNING: There are links to this heading in source code. If you change it, find and replace the links. -->

Rill uses the credentials configured in your local environment using the Google Cloud CLI (`gcloud`). Follow these steps to configure it:

1. Open a terminal window and run `gcloud auth list` to check if you already have the Google Cloud CLI installed and authenticated. 

2. If it did not print information about your user, follow the steps on [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk). Make sure to run `gcloud init` after installation as described in the tutorial.

Once you have `gcloud` installed, run this command to set your default credentials:

```
gcloud auth application-default login
```

You have now configured Google Cloud access from your local environment. Rill will detect and use your credentials next time you try to ingest a source.

### Configure credentials for S3

<!-- WARNING: There are links to this heading in source code. If you change it, find and replace the links. -->

Rill uses the credentials configured in your local environment using the AWS CLI. 

To check if you already have the AWS CLI installed and authenticated, open a terminal window and run:
```bash
aws iam get-user --no-cli-pager
```
> Note: The above command works with AWS CLI version 2 and above

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

### Configure credentials for Azure blob storage

<!-- WARNING: There are links to this heading in source code. If you change it, find and replace the links. -->

Rill uses the credentials configured in your local environment using the Azure CLI (`az`). Follow these steps to configure it:

1. Open a terminal window and run the following command to check if you already have the Azure CLI installed and authenticated:

    ```bash
    az account show
    ```

2. If it did not display information about your Azure account, you can [install the Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) if it is not already installed on your system.

3. After installing the Azure CLI, run the following command to authenticate with your Azure account:

    ```bash
    az login
    ```

    Follow the on-screen instructions to complete the login process.

    > If no web browser is available or the web browser fails to open, you may force device code flow with `az login --use-device-code`.

You have now configured Azure access from your local environment. Rill will detect and use your credentials the next time you interact with Azure services.

### Configure credentials for MotherDuck service

When developing a project locally, you need to set `motherduck_token` in your enviornment variables. Refer to motherduck [docs](https://motherduck.com/docs/authenticating-to-motherduck#saving-the-service-token-as-an-environment-variable) for more infromation on authenticating with token.
