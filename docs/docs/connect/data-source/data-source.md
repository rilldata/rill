---
title: "Connect to your Data"
description: Import local files or remote data sources
sidebar_label: "Connectors"
sidebar_position: 00
toc_max_heading_level: 3
className: connect-connect
---



import ConnectorIcon from '@site/src/components/ConnectorIcon';

Rill supports importing data from multiple sources to power your analytics and data applications. Whether you're working with cloud data warehouses, databases, file storage, or streaming data sources, Rill provides seamless connectivity and data ingestion capabilities.

## Getting Started with Connectors

In order to connect and browse through your data, you'll need to create a connector file. Browse through the options below for our supported connectors. Each connector is designed to handle the specific authentication and configuration requirements of your data source.

### What You Can Do After Connecting

Once the connector is created, you can directly create:
- **Models** - Transform and combine data from multiple sources using SQL
- **Metrics Views** - Define key business metrics and KPIs
- **Dashboards** - Visualize your data with interactive charts and tables

:::warning OLAP Engine Limitations
Rill supports connecting your data to both [DuckDB](/connect/olap/duckdb) and [ClickHouse](/connect/olap/clickhouse). However, there are still some features in development for managed ClickHouse. [Contact our team](/contact) for more informaton and scheduled feature releases!
:::


## Data Warehouse Connectors

### Athena
### BigQuery
### Redshift
### Snowflake

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Athena.svg" alt="Athena" />}
    header="Athena"
    content="Connect to Amazon Athena for serverless querying of data stored in S3 using standard SQL."
    link="/connect/data-source/athena"
    linkLabel="Learn more"
    referenceLink="athena"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Bigquery.svg" alt="BigQuery" />}
    header="BigQuery"
    content="Connect to Google BigQuery for analytics and data warehousing with service account authentication."
    link="/connect/data-source/bigquery"
    linkLabel="Learn more"
    referenceLink="bigquery"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Redshift.svg" alt="Redshift" />}
    header="Redshift"
    content="Connect to Amazon Redshift data warehouse with AWS credentials and support for both provisioned and serverless clusters."
    link="/connect/data-source/redshift"
    linkLabel="Learn more"
    referenceLink="redshift"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Snowflake.svg" alt="Snowflake" />}
    header="Snowflake"
    content="Connect to Snowflake data warehouse with support for individual credentials and JWT authentication."
    link="/connect/data-source/snowflake"
    linkLabel="Learn more"
    referenceLink="snowflake"
  />

</div>

## Databases
### MySQL
### PostgreSQL
### SQLite

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-mysql.svg" alt="MySQL" />}
    header="MySQL"
    content="Connect to MySQL databases with support for various authentication methods and SSL connections."
    link="/connect/data-source/mysql"
    linkLabel="Learn more"
    referenceLink="mysql"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Postgres.svg" alt="PostgreSQL" />}
    header="PostgreSQL"
    content="Connect to PostgreSQL databases with support for SSL connections and various authentication methods."
    link="/connect/data-source/postgres"
    linkLabel="Learn more"
    referenceLink="postgresql"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-SQLite.svg" alt="SQLite" />}
    header="SQLite"
    content="Connect to SQLite databases for lightweight, file-based data storage and querying."
    link="/connect/data-source/sqlite"
    linkLabel="Learn more"
    referenceLink="sqlite"
  />
</div>


## Cloud Storage Connectors

### Azure
### Google Cloud Storage
### Amazon S3

<div className="connector-icon-grid">

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-S3.svg" alt="Amazon S3" />}
    header="Amazon S3"
    content="Connect to Amazon S3 buckets to read data files including CSV, JSON, Parquet, and compressed formats."
    link="/connect/data-source/s3"
    linkLabel="Learn more"
    referenceLink="s3"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Azure.svg" alt="Microsoft Azure" />}
    header="Azure"
    content="Connect to Microsoft Azure Blob Storage to read data files with support for various formats."
    link="/connect/data-source/azure"
    linkLabel="Learn more"
    referenceLink="azure"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-GCS.svg" alt="Google Cloud Storage" />}
    header="Google Cloud Storage"
    content="Google Cloud Storage for scalable object storage and data lakes."
    link="/connect/data-source/gcs"
    linkLabel="Learn more"
    referenceLink="gcs"
  />


</div>

## Other Connectors

### HTTPS
### Local File
### Salesforce
### Google Sheets
### Slack

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<p className="https-icon">https:// </p>}
    header="HTTPS"
    content="Download data from HTTP/HTTPS URLs with support for various authentication methods."
    link="/connect/data-source/https"
    linkLabel="Learn more"
    referenceLink="https"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Local.svg" alt="Local File" />}
    header="Local File"
    content="Read data from local files including CSV, JSON, Parquet, and compressed formats."
    link="/connect/data-source/local-file"
    linkLabel="Learn more"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Salesforce.svg" alt="Salesforce" />}
    header="Salesforce"
    content="Connect to Salesforce to extract data from objects and queries using the Salesforce API."
    link="/connect/data-source/salesforce"
    linkLabel="Learn more"
    referenceLink="salesforce"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Sheets.svg" alt="Google Sheets" className="sheets-icon" />}
    header="Google Sheets"
    content="Connect to Google Sheets to read data from spreadsheets with support for multiple sheets."
    link="/connect/data-source/googlesheets"
    linkLabel="Learn more"
  />
</div>



## Externally Hosted Services
If you have a firewall in front of your externally hosted service, you will need to whitelist the IP addresses below. This will allow you to connect to/from your service once your project is deployed to Rill Cloud. 
```
35.196.245.100
34.74.117.37
35.196.153.31
34.75.22.143
34.148.167.51
35.237.60.193
```
**