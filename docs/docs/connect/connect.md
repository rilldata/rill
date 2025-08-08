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

- ### _[Rill Managed OLAP + Data Ingestion (Default)](/connect/data-source)_:
  
  Use Rill's embedded **ClickHouse / DuckDB** (depending on size of data) as the OLAP engine and ingest data from external sources.  Full Rill functionality with some caveats depending which embedded engine you select. 
 
      :::tip Rill Defaults with DuckDB
      When starting Rill for the first time, Rill will autopopulate the connector with a `duckdb.yaml`. To use ClickHouse, create a managed ClickHouse connector, `clickhouse.yaml` with `managed: true`. For more information, see our [ClickHouse](/connect/olap/clickhouse) docs.

      :::

- ### _[Bring Your Own OLAP (BYO OLAP)](/connect/olap)_: 
  
  Large-scale datasets (100GB+) or existing OLAP infrastructure Connect to existing **ClickHouse**, **Druid**, **Pinot**, or **MotherDuck** instances. Use Rill's "live connectors" to ingest data directly into your OLAP engines.

    :::note Modeling on BYO-OLAP
    Some modeling features may be limited depending on the engine.
    :::

## OLAP Engines

### DuckDB
### ClickHouse
### MotherDuck
### Druid
### Pinot

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-DuckDB.svg" alt="DuckDB" />}
    content="DuckDB is the default engine for Rill Developer."
    link="/connect/olap/duckdb"
    linkLabel="Learn more"
    referenceLink="duckdb"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Clickhouse.svg" alt="ClickHouse" />}
    content="High-performance columnar database for real-time analytics and data warehousing."
    link="/connect/olap/clickhouse"
    linkLabel="Learn more"
    referenceLink="clickhouse"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Motherduck.svg" alt="MotherDuck" />}
    content="Cloud-native DuckDB service for scalable analytics and data processing."
    link="/connect/olap/motherduck"
    linkLabel="Learn more"
    referenceLink="motherduck"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Druid.svg" alt="Druid" />}
    content="Real-time analytics database designed for high-performance OLAP queries."
    link="/connect/olap/druid"
    linkLabel="Learn more"
    referenceLink="druid"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Pinot.svg" alt="Pinot" />}
    content="Distributed OLAP datastore for real-time analytics and business intelligence."
    link="/connect/olap/pinot"
    linkLabel="Learn more"
    referenceLink="pinot"
  />
</div>

:::tip Missing an OLAP Engine?
Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, refer to our [OLAP Engines](/connect/olap) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](/contact)!
:::


## Data Warehouses

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


## Object Storage

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

## Other Data Connectors

### HTTPS
### Local File
### Salesforce


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
    content="Connect to public Google Sheets to read data from spreadsheets with support for multiple sheets."
    link="/connect/data-source/googlesheets"
    linkLabel="Learn more"
  />
</div>

:::tip Missing a connector?
We're constantly adding new data connectors. If you don't see what you need, [let us know](/contact) and we'll help you get connected.
:::

## Other Integrations

### Google Sheets
### Slack

<div className="connector-icon-grid">

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-Slack.svg" alt="Slack" className="sheets-icon" />}
    header="Slack"
    content="Connect to Slack to send alerts and messages from Rill."
    link="/connect/data-source/slack"
    linkLabel="Learn more"
    referenceLink="slack"
  />

  <ConnectorIcon
    icon={<img src="/img/connect/icons/Logo-AI.svg" alt="AI" className="sheets-icon" />}
    header="AI"
    content="Define your own OpenAI Connector and define your own API key."
    link="/build/metrics-view/#creating-metrics-with-ai"
    linkLabel="Learn more"
    referenceLink="ai"
  />
</div>

