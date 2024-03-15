---
title: Connectors
description: Connectors
sidebar_label: Connectors
sidebar_position: 0
---
## Overview

Rill supports connecting to a variety of data sources, including but not limited to object storage (S3, GCS, ABS), data warehouses (Snowflake, BigQuery), traditional RDBMS (Postgres, MySQL), and other analytics datastores (DuckDB / Motherduck, Athena, Salesforce, and more).

When running Rill locally, Rill Developer will establish a connection with existing credentials that have been configured on your computer (using embedded DuckDB). In Rill Cloud, a remote connection will be established using service account credentials that will need to be explicitly provided. For more information about available connectors and how to use them in Rill, please see the reference pages below.

:::tip Don't see a connector?

We are always adding new connectors as part of our release cycle. If there's a source of data that you'd like to ingest into Rill but don't see it, please don't hesitate to [contact us](contact.md) and give us your feedback!

:::

## List of Rill Connectors

- [Google Cloud Storage (GCS)](gcs.md)
- [Amazon S3](s3.md) 
- [Azure Blob Storage (Azure)](azure.md)
- [BigQuery](bigquery.md)
- [Amazon Athena](athena.md)
- [Amazon Redshift](redshift.md)
- [DuckDB / MotherDuck](motherduck.md)
- [PostgreSQL](postgres.md)
- [MySQL](mysql.md)
- [SQLite](sqlite.md)
- [Snowflake](snowflake.md)
- [Salesforce](salesforce.md)
- [Google Sheets](googlesheets.md)
