---
title: Connectors
description: Connectors
sidebar_label: Connectors
sidebar_position: 00
---
## Overview

Rill supports connecting to a variety of data sources, including but not limited to object storage (S3, GCS, ABS), data warehouses (Snowflake, BigQuery), traditional RDBMS (Postgres, MySQL), and other analytics datastores (DuckDB / Motherduck, Athena, Salesforce, and more).

When running Rill locally, Rill Developer will establish a connection with existing credentials that have been configured on your computer (using embedded DuckDB). In Rill Cloud, a remote connection will be established using service account credentials that will need to be explicitly provided. 

For more information about available connectors and how to use them in Rill (locally and in the cloud), please refer to the reference pages below.

## List of Rill Connectors

- [Amazon S3](s3.md) 
- [Google Cloud Storage (GCS)](gcs.md)
- [Azure Blob Storage (Azure)](azure.md)
- [Amazon Athena](athena.md)
- [BigQuery](bigquery.md)
- [MotherDuck](motherduck.md)
- [Postgres](postgres.md)
- [Salesforce](salesforce.md)
- [Snowflake](snowflake.md)
- [Google Sheets](googlesheets.md)
