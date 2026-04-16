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
import OLAPToggle from '@site/src/components/OLAPToggle';
import DuckDBLogo from '@site/static/img/build/connectors/icons/Logo-DuckDB.svg';
import ClickHouseLogo from '@site/static/img/build/connectors/icons/Logo-ClickHouse.svg';
import MongoDBLogo from '@site/static/img/build/connectors/icons/Logo-MongoDB.svg';
import MotherDuckLogo from '@site/static/img/build/connectors/icons/Logo-MotherDuck.svg';
import DruidLogo from '@site/static/img/build/connectors/icons/Logo-Druid.svg';
import StarRocksLogo from '@site/static/img/build/connectors/icons/Logo-StarRocks.svg';
import AthenaLogo from '@site/static/img/build/connectors/icons/Logo-Athena.svg';
import BigQueryLogo from '@site/static/img/build/connectors/icons/Logo-BigQuery.svg';
import RedshiftLogo from '@site/static/img/build/connectors/icons/Logo-Redshift.svg';
import SnowflakeLogo from '@site/static/img/build/connectors/icons/Logo-Snowflake.svg';
import MySQLLogo from '@site/static/img/build/connectors/icons/Logo-MySQL.svg';
import PostgresLogo from '@site/static/img/build/connectors/icons/Logo-Postgres.svg';
import SQLiteLogo from '@site/static/img/build/connectors/icons/Logo-SQLite.svg';
import SupabaseLogo from '@site/static/img/build/connectors/icons/Logo-Supabase.svg';
import S3Logo from '@site/static/img/build/connectors/icons/Logo-S3.svg';
import GCSLogo from '@site/static/img/build/connectors/icons/Logo-GCS.svg';
import AzureLogo from '@site/static/img/build/connectors/icons/Logo-Azure.svg';
import IcebergLogo from '@site/static/img/build/connectors/icons/Logo-Iceberg.svg';
import SheetsLogo from '@site/static/img/build/connectors/icons/Logo-Sheets.svg';
import LocalLogo from '@site/static/img/build/connectors/icons/Logo-Local.svg';
import SalesforceLogo from '@site/static/img/build/connectors/icons/Logo-Salesforce.svg';
import DeltaLakeLogo from '@site/static/img/build/connectors/icons/Logo-DeltaLake.svg';
import HadoopLogo from '@site/static/img/build/connectors/icons/Logo-Hadoop.svg';
import ClaudeLogo from '@site/static/img/build/connectors/icons/Logo-Claude.svg';
import GeminiLogo from '@site/static/img/build/connectors/icons/Logo-Gemini.svg';
import AILogo from '@site/static/img/build/connectors/icons/Logo-AI.svg';
import SlackLogo from '@site/static/img/build/connectors/icons/Logo-Slack.svg';
import KafkaLogo from '@site/static/img/build/connectors/icons/Logo-Kafka.svg';

## Connection Strategies

Rill offers flexible connection strategies to fit different data architectures and requirements.

- ### _[Rill Managed OLAP + Data Ingestion (Default)](/developers/build/connectors/data-source)_:
  
  Use Rill's embedded **ClickHouse / DuckDB** (depending on size of data) as the OLAP engine and ingest data from external sources. Full Rill functionality is available with [some caveats](/developers/build/connectors/data-source#managed-olap-engine-caveats) depending on which embedded engine you select.
 
      :::tip Rill Defaults with DuckDB
      When starting Rill for the first time, Rill will auto-populate the connector with a `duckdb.yaml`. To use ClickHouse, create a managed ClickHouse connector by selecting "Add Data", then ClickHouse, and finally "Rill-managed ClickHouse" in the UI. For more information, see [Rill Managed ClickHouse](/developers/build/connectors/olap/clickhouse#rill-managed-clickhouse).
      :::

- ### _[Bring Your Own OLAP (BYO OLAP)](/developers/build/connectors/olap)_: 
  
  For large-scale datasets (100GB+) or existing [OLAP infrastructure](/developers/build/connectors/olap#what-is-olap), connect to existing **ClickHouse**, **Druid**, **Pinot**, or **MotherDuck** instances. Use Rill's "live connectors" to ingest data directly into your OLAP engines.


## Rill Managed OLAP Engines
### ClickHouse
### DuckDB

Rill provisions and manages these engines for you — no infrastructure to set up.

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<ClickHouseLogo />}
    content="Rill's recommended engine for production. Managed ClickHouse handles large-scale datasets with high concurrency."
    link="/developers/build/connectors/olap/clickhouse"
    linkLabel="Learn more"
    referenceLink="clickhouse"
  />
  <ConnectorIcon
    icon={<DuckDBLogo />}
    content="The default engine for Rill Developer. Embedded and zero-config for local development."
    link="/developers/build/connectors/olap/duckdb"
    linkLabel="Learn more"
    referenceLink="duckdb"
  />
</div>

## Bring Your Own OLAP

Connect Rill to an existing OLAP engine you manage. Rill pushes queries down to your engine with no data ingestion.

### ClickHouse
### Druid
### MotherDuck
### Pinot
### Snowflake
### StarRocks

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<ClickHouseLogo />}
    content="Connect to your own self-managed ClickHouse or ClickHouse Cloud instance."
    link="/developers/build/connectors/olap/clickhouse"
    linkLabel="Learn more"
    referenceLink="clickhouse"
  />
  <ConnectorIcon
    icon={<DruidLogo />}
    content="Real-time analytics database designed for high-performance OLAP queries."
    link="/developers/build/connectors/olap/druid"
    linkLabel="Learn more"
    referenceLink="druid"
  />
  <ConnectorIcon
    icon={<MotherDuckLogo />}
    content="Cloud-native DuckDB service for scalable analytics and data processing."
    link="/developers/build/connectors/olap/motherduck"
    linkLabel="Learn more"
    referenceLink="motherduck"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Pinot.svg" alt="Pinot" />}
    content="Distributed OLAP datastore for real-time analytics and business intelligence."
    link="/developers/build/connectors/olap/pinot"
    linkLabel="Learn more"
    referenceLink="pinot"
  />
  <ConnectorIcon
    icon={<SnowflakeLogo />}
    content="Cloud data warehouse with native support for metrics views as a live connector."
    link="/developers/build/connectors/olap/snowflake"
    linkLabel="Learn more"
    referenceLink="snowflake"
  />
  <ConnectorIcon
    icon={<StarRocksLogo />}
    content="Distributed OLAP datastore for real-time analytics and business intelligence."
    link="/developers/build/connectors/olap/starrocks"
    linkLabel="Learn more"
    referenceLink="starrocks"
  />
</div>

:::tip Missing an OLAP Engine?
Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, refer to our [OLAP Engines](/developers/build/connectors/olap) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](/contact)!
:::

<OLAPToggle>
<OLAPToggle.DuckDB>

## Data Warehouses

### Athena
### BigQuery
### Redshift
### Snowflake

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<AthenaLogo />}
    header="Athena"
    content="Connect to Amazon Athena for serverless querying of data stored in S3 using standard SQL."
    link="/developers/build/connectors/data-source/duckdb/athena"
    linkLabel="Learn more"
    referenceLink="athena"
  />
  <ConnectorIcon
    icon={<BigQueryLogo />}
    header="BigQuery"
    content="Connect to Google BigQuery for analytics and data warehousing with service account authentication."
    link="/developers/build/connectors/data-source/duckdb/bigquery"
    linkLabel="Learn more"
    referenceLink="bigquery"
  />

  <ConnectorIcon
    icon={<RedshiftLogo />}
    header="Redshift"
    content="Connect to Amazon Redshift data warehouse with AWS credentials and support for both provisioned and serverless clusters."
    link="/developers/build/connectors/data-source/duckdb/redshift"
    linkLabel="Learn more"
    referenceLink="redshift"
  />
  <ConnectorIcon
    icon={<SnowflakeLogo />}
    header="Snowflake"
    content="Connect to Snowflake data warehouse with support for individual credentials and JWT authentication."
    link="/developers/build/connectors/data-source/duckdb/snowflake"
    linkLabel="Learn more"
    referenceLink="snowflake"
  />

</div>

## Databases
### MySQL
### PostgreSQL
### SQLite
### Supabase

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<MySQLLogo />}
    header="MySQL"
    content="Connect to MySQL databases with support for various authentication methods and SSL connections."
    link="/developers/build/connectors/data-source/duckdb/mysql"
    linkLabel="Learn more"
    referenceLink="mysql"
  />
  <ConnectorIcon
    icon={<PostgresLogo />}
    header="PostgreSQL"
    content="Connect to PostgreSQL databases with support for SSL connections and various authentication methods."
    link="/developers/build/connectors/data-source/duckdb/postgres"
    linkLabel="Learn more"
    referenceLink="postgresql"
  />
  <ConnectorIcon
    icon={<SQLiteLogo />}
    header="SQLite"
    content="Connect to SQLite databases for lightweight, file-based data storage and querying."
    link="/developers/build/connectors/data-source/duckdb/sqlite"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<SupabaseLogo />}
    header="Supabase"
    content="Connect to Supabase's managed PostgreSQL databases with SSL support and standard connection methods."
    link="/developers/build/connectors/data-source/duckdb/supabase"
    linkLabel="Learn more"
    referenceLink="supabase"
  />
</div>


## Object Storage

### Amazon S3
### Microsoft Azure Blob Storage
### Google Cloud Storage



<div className="connector-icon-grid">

  <ConnectorIcon
    icon={<S3Logo />}
    header="Amazon S3"
    content="Connect to Amazon S3 buckets to read data files including CSV, JSON, Parquet, and compressed formats."
    link="/developers/build/connectors/data-source/duckdb/s3"
    linkLabel="Learn more"
    referenceLink="s3"
  />
  <ConnectorIcon
    icon={<AzureLogo />}
    header="Azure"
    content="Connect to Microsoft Azure Blob Storage to read data files with support for various formats."
    link="/developers/build/connectors/data-source/duckdb/azure"
    linkLabel="Learn more"
    referenceLink="azure"
  />
  <ConnectorIcon
    icon={<GCSLogo />}
    header="Google Cloud Storage"
    content="Google Cloud Storage provides scalable object storage and data lakes."
    link="/developers/build/connectors/data-source/duckdb/gcs"
    linkLabel="Learn more"
    referenceLink="gcs"
  />

</div>

## Table Formats
### Delta Lake
### Apache Iceberg

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Delta.svg" alt="Delta Lake" />}
    header="Delta Lake"
    content="Read Delta tables directly from object storage through compatible query engines."
    link="/developers/build/connectors/data-source/duckdb/delta"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Iceberg.svg" alt="Apache Iceberg" />}
    header="Apache Iceberg"
    content="Read Iceberg tables directly from object storage through compatible query engines."
    link="/developers/build/connectors/data-source/duckdb/iceberg"
    linkLabel="Learn more"
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
    icon={<DuckDBLogo />}
    header="DuckDB"
    content="Attach your local DuckDB database to Rill's embedded database."
    link="/developers/build/connectors/data-source/duckdb/duckdb"
    linkLabel="Learn more"
    referenceLink="external-duckdb"
  />
  <ConnectorIcon
    icon={<SheetsLogo />}
    header="Google Sheets"
    content="Connect to public Google Sheets to read data from spreadsheets with support for multiple sheets."
    link="/developers/build/connectors/data-source/duckdb/googlesheets"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<p className="https-icon">https:// </p>}
    header="HTTPS"
    content="Download data from HTTP/HTTPS URLs with support for various authentication methods."
    link="/developers/build/connectors/data-source/duckdb/https"
    linkLabel="Learn more"
    referenceLink="https"
  />
  <ConnectorIcon
    icon={<LocalLogo />}
    header="Local File"
    content="Read data from local files including CSV, JSON, Parquet, and compressed formats."
    link="/developers/build/connectors/data-source/duckdb/local-file"
    linkLabel="Learn more"
  />

  <ConnectorIcon
    icon={<SalesforceLogo />}
    header="Salesforce"
    content="Connect to Salesforce to extract data from objects and queries using the Salesforce API."
    link="/developers/build/connectors/data-source/duckdb/salesforce"
    linkLabel="Learn more"
  />

</div>

:::tip Missing a connector?
We're constantly adding new data connectors. If you don't see what you need, [let us know](/contact) and we'll help you get connected.
:::

</OLAPToggle.DuckDB>
<OLAPToggle.ClickHouse>

<!-- ## Data Warehouses

:::note Staging Models

:::

### BigQuery
### Snowflake

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<BigQueryLogo />}
    header="BigQuery"
    content="Connect to Google BigQuery for analytics and data warehousing with service account authentication."
    link="/developers/build/connectors/data-source/duckdb/bigquery"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<SnowflakeLogo />}
    header="Snowflake"
    content="Connect to Snowflake data warehouse with support for individual credentials and JWT authentication."
    link="/developers/build/connectors/data-source/duckdb/snowflake"
    linkLabel="Learn more"
  />
</div> -->

## Databases
### MongoDB
### MySQL
### PostgreSQL
### Supabase

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<MongoDBLogo />}
    header="MongoDB"
    content="Connect to MongoDB collections using ClickHouse's mongodb() table function."
    link="/developers/build/connectors/data-source/clickhouse/mongodb"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<MySQLLogo />}
    header="MySQL"
    content="Connect to MySQL databases using ClickHouse's mysql() table function."
    link="/developers/build/connectors/data-source/clickhouse/mysql"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<PostgresLogo />}
    header="PostgreSQL"
    content="Connect to PostgreSQL databases using ClickHouse's postgresql() table function."
    link="/developers/build/connectors/data-source/clickhouse/postgres"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<SupabaseLogo />}
    header="Supabase"
    content="Connect to Supabase's managed PostgreSQL databases via ClickHouse's postgresql() table function."
    link="/developers/build/connectors/data-source/clickhouse/supabase"
    linkLabel="Learn more"
  />
</div>

## Object Storage

### Amazon S3
### Microsoft Azure Blob Storage
### Google Cloud Storage
### HDFS

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<S3Logo />}
    header="Amazon S3"
    content="Connect to Amazon S3 buckets using ClickHouse's s3() table function."
    link="/developers/build/connectors/data-source/clickhouse/s3"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<AzureLogo />}
    header="Azure"
    content="Connect to Azure Blob Storage using ClickHouse's azureBlobStorage() table function."
    link="/developers/build/connectors/data-source/clickhouse/azure"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<GCSLogo />}
    header="Google Cloud Storage"
    content="Connect to GCS using ClickHouse's gcs() table function with HMAC keys."
    link="/developers/build/connectors/data-source/clickhouse/gcs"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<HadoopLogo />}
    header="HDFS"
    content="Read data files from HDFS with support for Parquet, CSV, JSON, and other formats."
    link="/developers/build/connectors/data-source/clickhouse/hdfs"
    linkLabel="Learn more"
  />
</div>

## Table Formats
### Apache Hudi
### Apache Iceberg
### Delta Lake

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Hudi.png" alt="Apache Hudi" />}
    header="Apache Hudi"
    content="Read Hudi tables using ClickHouse's hudi() table function."
    link="/developers/build/connectors/data-source/clickhouse/hudi"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<IcebergLogo />}
    header="Apache Iceberg"
    content="Read Iceberg tables using ClickHouse's icebergS3() table function."
    link="/developers/build/connectors/data-source/clickhouse/iceberg"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<DeltaLakeLogo />}
    header="Delta Lake"
    content="Read Delta Lake tables using ClickHouse's deltaLake() table function."
    link="/developers/build/connectors/data-source/clickhouse/delta-lake"
    linkLabel="Learn more"
  />
</div>

## Other Data Connectors
### HTTPS
### Kafka
### Remote ClickHouse

<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<p className="https-icon">https:// </p>}
    header="HTTPS"
    content="Download data from HTTP/HTTPS URLs using ClickHouse's url() table function."
    link="/developers/build/connectors/data-source/clickhouse/https"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<KafkaLogo />}
    header="Kafka"
    content="Stream data from Kafka topics into ClickHouse using the Kafka table engine."
    link="/developers/build/connectors/data-source/clickhouse/kafka"
    linkLabel="Learn more"
  />
  <ConnectorIcon
    icon={<ClickHouseLogo />}
    header="Remote ClickHouse"
    content="Query data from other ClickHouse servers for cross-cluster analytics."
    link="/developers/build/connectors/data-source/clickhouse/remote-clickhouse"
    linkLabel="Learn more"
  />
</div>

:::tip Missing a connector?
We're constantly adding new data connectors. If you don't see what you need, [let us know](/contact) and we'll help you get connected.
:::

</OLAPToggle.ClickHouse>
</OLAPToggle>

## Service Integrations


### Claude
### Gemini
### OpenAI
### Slack


<div className="connector-icon-grid">
  <ConnectorIcon
    icon={<ClaudeLogo />}
    header="AI"
    content="Create and define a Claude Connector with your own API key."
    link="/developers/build/connectors/services/claude"
    linkLabel="Learn more"
    referenceLink="claude"
  />
  <ConnectorIcon
    icon={<GeminiLogo />}
    header="Gemini"
    content="Create and define a Gemini Connector with your own API key."
    link="/developers/build/connectors/services/gemini"
    linkLabel="Learn more"
    referenceLink="gemini"
  />
  <ConnectorIcon
    icon={<AILogo />}
    header="AI"
    content="Create and define an OpenAI Connector with your own API key."
    link="/developers/build/connectors/services/openai"
    linkLabel="Learn more"
    referenceLink="openai"
  />

  <ConnectorIcon
    icon={<SlackLogo />}
    header="Slack"
    content="Connect to Slack to send alerts and messages from Rill."
    link="/developers/build/connectors/services/slack"
    linkLabel="Learn more"
    referenceLink="slack"
  />
</div>


