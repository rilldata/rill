---
title: BigQuery
sidebar_label: BigQuery
sidebar_position: 10
hide_table_of_contents: true
---

## Overview

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets using SQL-like queries. It supports a highly scalable and flexible architecture, allowing users to analyze large amounts of data in real time, making it suitable for BI/ML applications. Rill supports natively connecting to and reading from BigQuery as a source by leveraging the [BigQuery SDK](https://cloud.google.com/bigquery/docs/reference/libraries).

## Local Credentials

Rill can connect to BigQuery using [Application Default Credentials (ADC)](https://cloud.google.com/docs/authentication/provide-credentials-adc#local-dev) credentials that have been set locally either in an environment variable (`GOOGLE_APPLICATION_CREDENTIALS`) or on a path already checked by the BigQuery SDK. In most cases, you should be able to run the following command to locally authenticate via ADC (which will set the credentials in a well-known location already checked by the BigQuery SDK):
```
gcloud auth application-default login
```

For more details, please refer to Google's official [Application Default Credentials](https://cloud.google.com/docs/authentication/provide-credentials-adc#local-dev) documentation.

:::info Deploying to Rill Cloud

Please note that when deploying a project to Rill Cloud, you will need to explicitly set and pass credentials during the deployment command / workflow. For more information, please see our [Credentials](../../deploy/credentials.md) documentation.

:::

## Configuring BigQuery as a source

Once BigQuery credentials have been configured and set locally, you should now be able to connect to and reference BigQuery tables in Rill! Under the hood, this is powered by using DuckDB's native BigQuery integration (which _does not require_ you to install any additional extensions).

Once everything has been set up, you will be able to query BigQuery tables directly using DuckDB SQL and table functions. Depending on how you'd like to set up and/or configure the source to ingest from BigQuery, there are generally 2 separate methods that can be used _(the second being more configurable than the first)_:

### Method 1 - Using BigQuery with DuckDB's SQL SELECT
Instead of creating a source YAML file, you can directly query the table from a model SQL file!

```sql
-- If you want to use BigQuery directly in a model,
-- you can directly query the source table via DuckDB SQL.
-- NOTE - You should have already authenticated locally using `gcloud auth application-default login`

SELECT * FROM bigquery_scan('my_gcp_project.my_dataset.source_table');
```

### Method 2 - Creating a source YAML file

Alternatively, if you'd like more control to configure and/or document the underlying data source, you can create a _source YAML file_. Create a new `<name>.yaml` source file in your Rill project directory (in our example, under `sources`) and enter the following:

```yaml
# BigQuery Source YAML

# Type will always be "source" for sources
type: source

# Connector to use - for BigQuery, we can use the duckdb connector
connector: duckdb

# The SQL query that Rill will run to ingest the data into DuckDB
sql: SELECT * FROM bigquery_scan('my_gcp_project.my_dataset.source_table');

# This is commented out but for reference, you may optionally add a trigger to refresh the source on a schedule
# refresh:
#   cron: "0 */6 * * *" # Refresh every 6 hours (optional)
```

:::info

For the definition / list of available properties for sources, please refer to [source YAML](../../reference/project-files/sources.md).

:::

:::tip Did you know?

If you have a BigQuery table name with special characters, you can escape these special characters in DuckDB by wrapping them in double quotes. For example, `bigquery_scan('my_gcp_project.my_dataset."data.table"')`. For more details, please see our [DuckDB identifier docs](https://duckdb.org/docs/sql/dialect/keywords_and_identifiers.html#rules-for-case-sensitivity).

:::

## Deploying to Rill Cloud

Once you have modeled your data with Rill locally using BigQuery as a source and you are ready to deploy your project to Rill Cloud, you will need to explicitly set credentials that can be used by Rill Cloud to connect to BigQuery.

### Using Service Account credentials in Rill Cloud

Rather than deploying with your personal credentials, we recommended creating and using a [Service Account](https://cloud.google.com/iam/docs/service-account-overview) for deployment to Rill Cloud. Google Service Accounts represent non-human identities in the Google Cloud and can be assigned custom permissions and access to certain resources within the GCP (including specific datasets or tables in BigQuery).

In the command line, you can run `rill env configure` to set credentials that Rill Cloud can use to access BigQuery:
```
rill env configure
```
You will be prompted to enter a [base64-encoded service account JSON key](https://cloud.google.com/iam/docs/keys-create-delete#creating) in a multi-line input (press Return + Ctrl+D when done entering the key):
```
$ rill env configure
? Which connector would you like to configure? bigquery
? Enter your BigQuery credentials [type: string, hint: gcp service account json]: Paste in the base64 encoded version here!
```

:::info
For more details on creating a Service Account and base64-encoding a JSON key, please check out our [Deploy with BigQuery credentials](../../deploy/credentials.md#bigquery) documentation.
:::

Then, run the deploy command and follow the prompts to deploy your project to Rill Cloud:
```
rill deploy
```
