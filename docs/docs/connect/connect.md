---
title: "Connect to your Data"
description: Import local files or remote data sources
sidebar_label: "Connectors"
sidebar_position: 00
toc_max_heading_level: 3
className: connect-connect
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

import ConnectorIcon from '@site/src/components/ConnectorIcon';


## Connection Strategies

Rill offers flexible connection strategies to fit different data architectures and requirements.

- _[**Embedded OLAP + Data Ingestion (Default)**](#data-warehouse-connectors)_: Most use cases, datasets up to ~50GB

Use Rill's embedded **DuckDB** as the OLAP engine and ingest data from external sources. Full Rill functionality with excellent performance for smaller datasets.


- _[**Bring Your Own OLAP (BYO OLAP)**](/connect/olap/)_: Large-scale datasets (100GB+) or existing OLAP infrastructure

Connect to existing **ClickHouse**, **Druid**, **Pinot**, or **MotherDuck** instances. Use Rill's connectors to ingest data directly into your OLAP engine.

:::note Modeling on BYO OLAP
 Some modeling features may be limited depending on the engine.
:::

-  _[**BYO OLAP with Native Connectors**](/connect/olap/)_: Working with existing optimized tables

Skip data ingestion and work directly with existing tables in your OLAP engine, **ClickHouse**, **Druid**, **Pinot**, or **MotherDuck**. Leverage engine-specific features and avoid data duplication.

## Data Warehouse Connectors

### Athena
### BigQuery
### MySQL
### PostgreSQL
### Redshift
### Snowflake
### SQLite

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Athena.png" alt="Athena" />}
    header="Athena"
    content="Connect to Amazon Athena for serverless querying of data stored in S3 using standard SQL."
    link="/connect/connector/athena"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#athena"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Bigquery.png" alt="BigQuery" />}
    header="BigQuery"
    content="Connect to Google BigQuery for analytics and data warehousing with service account authentication."
    link="/connect/connector/bigquery"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#bigquery"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-mysql.png" alt="MySQL" />}
    header="MySQL"
    content="Connect to MySQL databases with support for various authentication methods and SSL connections."
    link="/connect/connector/mysql"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#mysql"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Postgres.png" alt="PostgreSQL" />}
    header="PostgreSQL"
    content="Connect to PostgreSQL databases with support for SSL connections and various authentication methods."
    link="/connect/connector/postgres"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#postgres"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Redshift.png" alt="Redshift" />}
    header="Redshift"
    content="Connect to Amazon Redshift data warehouse with AWS credentials and support for both provisioned and serverless clusters."
    link="/connect/connector/redshift"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#redshift"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Snowflake.png" alt="Snowflake" />}
    header="Snowflake"
    content="Connect to Snowflake data warehouse with support for individual credentials and JWT authentication."
    link="/connect/connector/snowflake"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#snowflake"
  />
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-SQLite.png" alt="SQLite" />}
    header="SQLite"
    content="Connect to SQLite databases for lightweight, file-based data storage and querying."
    link="/connect/connector/sqlite"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#sqlite"
  />
</div>

## Cloud Storage Connectors

### Azure
### Google Cloud Storage
### Amazon S3

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Azure.png" alt="Microsoft Azure" />}
    header="Azure"
    content="Connect to Microsoft Azure Blob Storage to read data files with support for various formats."
    link="/connect/connector/azure"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#azure"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-GCS.png" alt="Google Cloud Storage" />}
    header="Google Cloud Storage"
    content="Google Cloud Storage for scalable object storage and data lakes."
    link="/connect/connector/gcs"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#gcs"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-S3.png" alt="Amazon S3" />}
    header="Amazon S3"
    content="Connect to Amazon S3 buckets to read data files including CSV, JSON, Parquet, and compressed formats."
    link="/connect/connector/s3"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#s3"
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
    link="/connect/connector/https"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#https"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Local.png" alt="Local File" />}
    header="Local File"
    content="Read data from local files including CSV, JSON, Parquet, and compressed formats."
    link="/connect/connector/local-file"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#local-file"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Salesforce.png" alt="Salesforce" />}
    header="Salesforce"
    content="Connect to Salesforce to extract data from objects and queries using the Salesforce API."
    link="/connect/connector/salesforce"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#salesforce"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Sheets.png" alt="Google Sheets" className="sheets-icon" />}
    header="Google Sheets"
    content="Connect to Google Sheets to read data from spreadsheets with support for multiple sheets."
    link="/connect/connector/googlesheets"
    linkLabel="Learn more"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Slack.png" alt="Slack" className="sheets-icon" />}
    header="Slack"
    content="Connect to Slack to extract data from channels, messages, and other workspace information."
    link="/connect/connector/slack"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#slack"
  />
</div>

:::info Full List Of Connectors

Rill is continually adding new sources and connectors in our releases. For a comprehensive list, you can refer to our [Connectors](/connect/connector/) page. Please don't hesitate to [reach out](/contact) either if there's a connector you'd like us to add!

:::

:::tip Avoid Pre-aggregated Metrics

Rill works best for slicing and dicing data meaning keeping data closer to raw to retain that granularity for flexible analysis. When loading data, be careful with adding pre-aggregated metrics like averages as that could lead to unintended results like a sum of an average. Instead, load the two raw metrics and calculate the derived metric in your model or dashboard.

:::

:::note Have a firewall setup?
You need to whitelist the following IP addresses to connect to/from Rill Cloud and your service behind the firewall.
```
35.196.245.100
34.74.117.37
35.196.153.31
34.75.22.143
34.148.167.51
35.237.60.193
```
:::
