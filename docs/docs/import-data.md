---
title: Import data source
description: Import local files, remote sources, and existing DuckDB databases
---

There are several ways to import data in Rill Developer: local files, remotely stored files, and existing DuckDB databases.

## Local files
To import a file using the UI, click "+" by Sources in the left hand navigation pane, select "File", and navigate to the specific file. Alternately, try dragging and dropping the file directly onto the Rill interface.

To import a file with the CLI use the [`rill import-source`](/cli#import-your-data) CLI command from the terminal.

```
rill import-source /path/to/data_1.parquet
rill import-source /path/to/data_2.csv
rill import-source /path/to/data_3.tsv
```

## Remote sources
To add a remote source using the UI, click "+" by Sources in the left hand navigation pane and select the location where your files are stored ("Google Cloud Storage" or "Amazon S3"). Enter your file's URI and source name before clicking "Add Source".

To access private data, you'll need to configure your local machine with credentials to the relevant cloud provider (see instructions below). Rill uses official cloud SDKs to automatically detect your local credentials and pass them on to the cloud platform. Your credentials are never stored in Rill.

After import, you'll be able to reimport your data whenever you want by clicking the "refresh source" button in the Rill UI.

Creating remote sources is not currently available through the CLI.

### Setting Google GCS credentials
Google Cloud Platform credentials are enabled through `gcloud` authentication in the terminal.

First, ensure you have the `gcloud` CLI installed locally by running the following CLI command. If it is not installed, go through the [gcloud install steps](https://cloud.google.com/sdk/docs/install).

```
gcloud --version
```

Next, authenticate your local machine. The following command opens a browser window and takes you through the Google authentication flow:

```
gcloud auth application-default login
```

Upon login, private GCS files available to this account can be pulled into Rill.


### Setting Amazon S3 credentials
You can use an access key and access secret to ingest privately stored data from S3 into Rill. This can be configured using the CLI.

First, ensure you have the AWS CLI installed locally by running the following command. If it is not installed, go through the [AWS CLI install steps](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html).

```
aws --version
```

Next, create an [access key and secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html) for a user that has _S3 read access_. Use the information from this account to configure your local AWS credentials:

```
aws configure
```

Enter the Access Key, Access Secret, and region for the AWS user with S3 read access. The default output format has no effect on Rill.

```
AWS Access Key ID [None]]: <your secret access ID>
AWS Secret Access Key [None]: <your secret access key>
Default region name [None]: <your region>
Default output format [None]: <None>
```

Private S3 files available to this account can now be pulled into Rill.

## Existing DuckDB databases

### Connecting
You can connect to an existing DuckDB database by running `rill init` and passing the `--db` option with a path to the db file.

Any updates made directly to the sources in the database will be reflected in Rill Developer. Similarly, any changes made by Rill Developer will modify the database.

Make sure to have only one connection open to the database, otherwise there will be some unexpected issues.

```
rill init --db /path/to/duckdb/database.db
```

### Copying
You can also copy over the database so that there are no conflicts and overrides that are propagated to the source by passing the `--db` option with `--copy` to achieve this.

```
rill init --db /path/to/duckdb/database.db --copy
```

## Limitations

Today, a few constraints apply to the data sources you can import:
- Only Parquet and CSV files are supported.
- gzipped files are not yet supported.
- You can only import a single file at a time.

Look out for gzipped file support & glob pattern support in Rill's next release!

## Request a new connector
If you don't see your data source listed above, [please let us know](https://discord.gg/eEvSYHdfWK)! We're continually adding new connectors, so your feedback will help us prioritize what data sources to support next.
