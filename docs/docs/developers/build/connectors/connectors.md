---
title: "Connect to your Data"
description: Import local files or remote data sources
sidebar_label: "Connectors"
sidebar_position: 0
toc_max_heading_level: 3
className: connect-connect
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

import ConnectorIcon from '@site/src/components/ConnectorIcon';

## Connection Strategies

Rill offers flexible connection strategies to fit different data architectures and requirements.

- ### _[Rill Managed OLAP + Data Ingestion (Default)](/developers/build/connectors/data-source)_:
  
  Use Rill's embedded **ClickHouse / DuckDB** (depending on size of data) as the OLAP engine and ingest data from external sources. Full Rill functionality is available with [some caveats](/developers/build/connectors/data-source#managed-olap-engine-caveats) depending on which embedded engine you select.
 
      :::tip Rill Defaults with DuckDB
      When starting Rill for the first time, Rill will auto-populate the connector with a `duckdb.yaml`. To use ClickHouse, create a managed ClickHouse connector by selecting "Add Data", then ClickHouse, and finally "Rill-managed ClickHouse" in the UI. For more information, see [Rill Managed ClickHouse](/developers/build/connectors/olap/clickhouse#rill-managed-clickhouse).
      :::

- ### _[Bring Your Own OLAP (BYO OLAP)](/developers/build/connectors/olap)_: 
  
  For large-scale datasets (100GB+) or existing [OLAP infrastructure](/developers/build/connectors/olap#what-is-olap), connect to existing **ClickHouse**, **Druid**, **Pinot**, or **MotherDuck** instances. Use Rill's "live connectors" to ingest data directly into your OLAP engines.

## OLAP Engines

### DuckDB
### ClickHouse
### MotherDuck
### Druid
### Pinot
### StarRocks

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-DuckDB.svg" alt="DuckDB" />}
    content="DuckDB is the default engine for Rill Developer."
    link="/developers/build/connectors/olap/duckdb"
    linkLabel="Learn more"
    referenceLink="duckdb"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-ClickHouse.svg" alt="ClickHouse" />}
    content="High-performance columnar database for real-time analytics and data warehousing."
    link="/developers/build/connectors/olap/clickhouse"
    linkLabel="Learn more"
    referenceLink="clickhouse"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-MotherDuck.svg" alt="MotherDuck" />}
    content="Cloud-native DuckDB service for scalable analytics and data processing."
    link="/developers/build/connectors/olap/motherduck"
    linkLabel="Learn more"
    referenceLink="motherduck"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Druid.svg" alt="Druid" />}
    content="Real-time analytics database designed for high-performance OLAP queries."
    link="/developers/build/connectors/olap/druid"
    linkLabel="Learn more"
    referenceLink="druid"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Pinot.svg" alt="Pinot" />}
    content="Distributed OLAP datastore for real-time analytics and business intelligence."
    link="/developers/build/connectors/olap/pinot"
    linkLabel="Learn more"
    referenceLink="pinot"
  />
  
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-StarRocks.png" alt="StarRocks" className="starrocks-icon"/>}
    content="Distributed OLAP datastore for real-time analytics and business intelligence."
    link="/developers/build/connectors/olap/starrocks"
    linkLabel="Learn more"
    referenceLink="starrocks"
  />
</div>

:::tip Missing an OLAP Engine?
Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, refer to our [OLAP Engines](/developers/build/connectors/olap) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](/contact)!
:::


## Data Warehouses

### Athena
### BigQuery
### Redshift
### Snowflake

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Athena.svg" alt="Athena" />}
    header="Athena"
    content="Connect to Amazon Athena for serverless querying of data stored in S3 using standard SQL."
    link="/developers/build/connectors/data-source/athena"
    linkLabel="Learn more"
    referenceLink="athena"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-BigQuery.svg" alt="BigQuery" />}
    header="BigQuery"
    content="Connect to Google BigQuery for analytics and data warehousing with service account authentication."
    link="/developers/build/connectors/data-source/bigquery"
    linkLabel="Learn more"
    referenceLink="bigquery"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Redshift.svg" alt="Redshift" />}
    header="Redshift"
    content="Connect to Amazon Redshift data warehouse with AWS credentials and support for both provisioned and serverless clusters."
    link="/developers/build/connectors/data-source/redshift"
    linkLabel="Learn more"
    referenceLink="redshift"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Snowflake.svg" alt="Snowflake" />}
    header="Snowflake"
    content="Connect to Snowflake data warehouse with support for individual credentials and JWT authentication."
    link="/developers/build/connectors/data-source/snowflake"
    linkLabel="Learn more"
    referenceLink="snowflake"
  />

</div>

## Databases
### MySQL
### Oracle
### PostgreSQL
### SQLite

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-MySQL.svg" alt="MySQL" />}
    header="MySQL"
    content="Connect to MySQL databases with support for various authentication methods and SSL connections."
    link="/developers/build/connectors/data-source/mysql"
    linkLabel="Learn more"
    referenceLink="mysql"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Oracle.svg" alt="Oracle" />}
    header="Oracle"
    content="Connect to Oracle databases with support for service name and connection string authentication."
    link="/developers/build/connectors/data-source/oracle"
    linkLabel="Learn more"
    referenceLink="oracle"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Postgres.svg" alt="PostgreSQL" />}
    header="PostgreSQL"
    content="Connect to PostgreSQL databases with support for SSL connections and various authentication methods."
    link="/developers/build/connectors/data-source/postgres"
    linkLabel="Learn more"
    referenceLink="postgresql"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-SQLite.svg" alt="SQLite" />}
    header="SQLite"
    content="Connect to SQLite databases for lightweight, file-based data storage and querying."
    link="/developers/build/connectors/data-source/sqlite"
    linkLabel="Learn more"
    referenceLink="sqlite"
  />
</div>


## Object Storage

### Amazon S3
### Google Cloud Storage
### Microsoft Azure Blob Storage



<div className="connector-icon-grid">

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-S3.svg" alt="Amazon S3" />}
    header="Amazon S3"
    content="Connect to Amazon S3 buckets to read data files including CSV, JSON, Parquet, and compressed formats."
    link="/developers/build/connectors/data-source/s3"
    linkLabel="Learn more"
    referenceLink="s3"
  />
    <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-GCS.svg" alt="Google Cloud Storage" />}
    header="Google Cloud Storage"
    content="Google Cloud Storage provides scalable object storage and data lakes."
    link="/developers/build/connectors/data-source/gcs"
    linkLabel="Learn more"
    referenceLink="gcs"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Azure.svg" alt="Microsoft Azure" />}
    header="Azure"
    content="Connect to Microsoft Azure Blob Storage to read data files with support for various formats."
    link="/developers/build/connectors/data-source/azure"
    linkLabel="Learn more"
    referenceLink="azure"
  />


</div>

## Other Data Connectors
### External DuckDB
### Google Sheets
### HTTPS
### Local File
### Salesforce


<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-DuckDB.svg" alt="DuckDB" className="duckdb-icon"/>}
    header="DuckDB"
    content="Attach your local DuckDB database to Rill's embedded database."
    link="/developers/build/connectors/data-source/duckdb"
    linkLabel="Learn more"
    referenceLink="external-duckdb"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Sheets.svg" alt="Google Sheets" className="sheets-icon" />}
    header="Google Sheets"
    content="Connect to public Google Sheets to read data from spreadsheets with support for multiple sheets."
    link="/developers/build/connectors/data-source/googlesheets"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<p className="https-icon">https:// </p>}
    header="HTTPS"
    content="Download data from HTTP/HTTPS URLs with support for various authentication methods."
    link="/developers/build/connectors/data-source/https"
    linkLabel="Learn more"
    referenceLink="https"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Local.svg" alt="Local File" />}
    header="Local File"
    content="Read data from local files including CSV, JSON, Parquet, and compressed formats."
    link="/developers/build/connectors/data-source/local-file"
    linkLabel="Learn more"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Salesforce.svg" alt="Salesforce" />}
    header="Salesforce"
    content="Connect to Salesforce to extract data from objects and queries using the Salesforce API."
    link="/developers/build/connectors/data-source/salesforce"
    linkLabel="Learn more"
    referenceLink="salesforce"
  />

</div>

:::tip Missing a connector?
We're constantly adding new data connectors. If you don't see what you need, [let us know](/contact) and we'll help you get connected.
:::

## Service Integrations


### Claude
### Gemini
### OpenAI
### Slack


<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Claude.svg" alt="Claude" className="sheets-icon" />}
    header="AI"
    content="Create and define a Claude Connector with your own API key."
    link="/developers/build/connectors/services/claude"
    linkLabel="Learn more"
    referenceLink="claude"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Gemini.svg" alt="Gemini" className="sheets-icon" />}
    header="Gemini"
    content="Create and define a Gemini Connector with your own API key."
    link="/developers/build/connectors/services/gemini"
    linkLabel="Learn more"
    referenceLink="gemini"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-AI.svg" alt="OpenAI" className="sheets-icon" />}
    header="AI"
    content="Create and define an OpenAI Connector with your own API key."
    link="/developers/build/connectors/services/openai"
    linkLabel="Learn more"
    referenceLink="openai"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Slack.svg" alt="Slack" className="slack-icon" />}
    header="Slack"
    content="Connect to Slack to send alerts and messages from Rill."
    link="/developers/build/connectors/services/slack"
    linkLabel="Learn more"
    referenceLink="slack"
  />
</div>


