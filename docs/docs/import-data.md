---
title: Import data source
description: Import local files, remote sources, and existing DuckDB databases
---

There are several ways to import data in Rill Developer: local files, remotely stored files, and existing DuckDB databases.

A few things to note:
- Only Parquet and CSV files are supported.
- You can only import a single file into a source at a time.

## Local files
To import a file using the UI, click "+" by Sources in the left hand navigation pane, select "File", and navigate to the specific file. Alternately, try dragging and dropping the file directly onto the Rill interface.

To import a file with the CLI use the [`rill import-source`](/cli#import-your-data) CLI command from the terminal.

```
rill import-source /path/to/data_1.parquet
rill import-source /path/to/data_2.csv
rill import-source /path/to/data_3.tsv
```

## Remote sources
To add a remote source using the UI, click "+" by Sources in the left hand navigation pane and select the location where your files are stored ("Google GCS" or "Amazon S3"). Enter your file's URI and source name before clicking "Add Source". The remotely stored file can now be reimported with the click of a button.

- Accessing public data doesn't require any credentials. 
- Accessing private data requires valid credentials. Rill automatically detect the credentials set on your local environment to pull private data to into Rill sources. 

Creating remote sources is not currently available through the CLI.

### Setting Google GCS credentials
Google Cloud Platform credentials are enabled through gcloud authentication in the terminal.

Ensure you have the gcloud CLI installed locally. If it is not installed, go through the [gcloud install steps](https://cloud.google.com/sdk/docs/install).

```
gcloud --version
```

Authenticate your local machine by running this CLI command. This command opens a browser window and takes you through the Google authentication flow. Once you have logged in, private GCS files available to this account can be pulled into Rill.

```
gcloud auth application-default login
```


### Setting Amazon S3 credentials
You can use an access key and access secret to ingest privately stored data from S3 into Rill. This can be configured using the CLI.

Ensure you have the AWS CLI installed locally by running. If it is not installed, go through the [AWS CLI install steps](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html).

```
aws --version
```


Create an [access key and secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html) for a user that _has S3 read access_. Use the information from this account to configure your local aws credentials.

```
aws configure
```

Enter the Access Key, Access Secret, and region for the AWS user with S3 read access. The default output format can be set to `None`. Once the credentials are configured locally, private S3 files available to this account can be pulled into Rill.

```
AWS Access Key ID [None]]: <your secret access ID>
AWS Secret Access Key [None]: <your secret access key>
Default region name [None]: <your region>
Default output format [None]: <None>
```

## Existing DuckDB databases

### Connecting
You can connect to an existing DuckDB database by passing the `--db` option with a path to the db file.

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

## Request a new connector
If you don't see your data source listed above, [please let us know](https://discord.gg/eEvSYHdfWK)! We're continually adding new connectors, so your feedback will help us prioritize what data sources to support next.
