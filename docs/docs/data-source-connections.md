---
title: Data Source Connections
---

## Local files
Click "+" by `sources` on the left hand navigation and select "File" to ingest a local CSV or Parquet file as a new source. 

You can also use the [`rill import-source`](/cli#import-your-data) CLI command from the terminal.

```
rill import-source /path/to/data_1.parquet
rill import-source /path/to/data_2.csv
rill import-source /path/to/data_3.tsv
```

## Remote sources
Click "+" by `sources` on the left hand navigation and select the data lake where your files are stored (GCS or S3). Enter your source configuration and enjoy a refreshable update to your remotely stored file with the click of a button.

    A few things to note:
    - Only Parquet and CSV files are supported.
    - You can only import a single file at a time.
    - Public buckets don't require any authentication, but private buckets require credentials.


### Google Cloud Storage (GCS)

#### Accessing private storage with credentials
Rill will automatically to detect local Google Cloud credentials to bring private data into your local sources. This is enabled through gcloud authentication in the terminal.

Ensure you have the gcloud CLI installed locally.

```
gcloud --version
```

If it is not installed, go through the [gcloud install steps](https://cloud.google.com/sdk/docs/install).

Authenticate your local machine by running this CLI command: 

```
gcloud auth application-default login
```

Private GCS stored files available to this account should now be available to be added as sources in Rill

### AWS S3

#### Region 
You need to provide the region in which the bucket is located to bring down the data. Look in the details of your S3 bucket info to determine the region.

#### Accessing private storage with credentials
You can use an access key and access secret to ingest privately stored data from S3 into Rill. This can be done by using the CLI.

Ensure you have the AWS CLI installed locally by running:

```
aws --version
```

If it is not installed, go through the [AWS CLI install steps](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html).

Create an [access key and secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html) for a user that has S3 read access.

```
aws iam create-access-key
```

Enter the Access Key and Access Secret into the configuration modal and the file should now be available to add as a source to Rill.


## Existing DuckDB databases

### Connecting
You can connect to an existing DuckDB database by passing the `--db` option with a path to the db file.

Any updates made directly to the sources in the database will be reflected in Rill Developer.  Similarly, any changes made by Rill Developer will modify the database.

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
