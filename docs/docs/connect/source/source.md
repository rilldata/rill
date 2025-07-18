---
title: "Connect to your Data"
description: Import local files or remote data sources
sidebar_label: "Connectors"
sidebar_position: 00
---
<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

import TileIcon from '@site/src/components/TileIcon';

Rill supports a multitude of connectors to ingest data from various sources: local files, S3 or GCS buckets, downloads using HTTP(S), databases, data warehouses, and more. Rill supports ingestion of `.csv`, `.tsv`, `.json`, and `.parquet` files, including compressed versions (`.gz`). This can be done either through the UI directly, when working with Rill Developer, or by pushing the logic into the [source YAML](/reference/project-files/sources) definition directly (see _Using Code_ sections below).

### Data Warehouse Connectors

<div className="connector-tile-grid">
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Snowflake.png" alt="Snowflake" />}

    content="Connect to Snowflake data warehouse with support for individual credentials and JWT authentication."
    link="/connect/source/connectors/snowflake"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#snowflake"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Bigquery.png" alt="BigQuery" />}

    content="Connect to Google BigQuery for analytics and data warehousing with service account authentication."
    link="/connect/source/connectors/bigquery"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#bigquery"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Redshift.png" alt="Redshift" />}

    content="Connect to Amazon Redshift data warehouse with AWS credentials and support for both provisioned and serverless clusters."
    link="/connect/source/connectors/redshift"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#redshift"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Postgres.png" alt="PostgreSQL" />}

    content="Connect to PostgreSQL databases with support for SSL connections and various authentication methods."
    link="/connect/source/connectors/postgres"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#postgres"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Athena.png" alt="Athena" />}

    content="Connect to Amazon Athena for serverless querying of data stored in S3 using standard SQL."
    link="/connect/source/connectors/athena"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#athena"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-mysql.png" alt="MySQL" />}

    content="Connect to MySQL databases with support for various authentication methods and SSL connections."
    link="/connect/source/connectors/mysql"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#mysql"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-SQLite.png" alt="SQLite" />}

    content="Connect to SQLite databases for lightweight, file-based data storage and querying."
    link="/connect/source/connectors/sqlite"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#sqlite"
  />
  </div>
  ### Cloud Storage Connectors
  <div className="connector-tile-grid">
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-S3.png" alt="Amazon S3" />}

    content="Connect to Amazon S3 buckets to read data files including CSV, JSON, Parquet, and compressed formats."
    link="/connect/source/connectors/s3"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#s3"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-GCS.png" alt="Google Cloud Storage" />}

    content="Google Cloud Storage for scalable object storage and data lakes."
    link="/connect/source/connectors/gcs"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#gcs"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Azure.png" alt="Microsoft Azure" />}

    content="Connect to Microsoft Azure Blob Storage to read data files with support for various formats."
    link="/connect/source/connectors/azure"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#azure"
  />
  </div>

  ### Other Connectors
  <div className="connector-tile-grid">
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Salesforce.png" alt="Salesforce" />}

    content="Connect to Salesforce to extract data from objects and queries using the Salesforce API."
    link="/connect/source/connectors/salesforce"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#salesforce"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Sheets.png" alt="Google Sheets" className="sheets-icon" />}

    content="Connect to Google Sheets to read data from spreadsheets with support for multiple sheets."
    link="/connect/source/connectors/sheets"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#googlesheets"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Slack.png" alt="Slack" className="sheets-icon" />}

    content="Connect to Slack to extract data from channels, messages, and other workspace information."
    link="/connect/source/connectors/slack"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#slack"
  />
  
  <TileIcon
    icon={<img src="/img/connect/icons/Logo-Local.png" alt="Local File" />}

    content="Read data from local files including CSV, JSON, Parquet, and compressed formats."
    link="/connect/source/connectors/local-file"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#local-file"
  />
  
  <TileIcon
    icon={<p className="https-icon">https:// </p>}

    content="Download data from HTTP/HTTPS URLs with support for various authentication methods."
    link="/connect/source/connectors/https"
    linkLabel="Learn more"
    referenceLink="/reference/project-files/connectors#https"
  />
</div>

:::info Full List Of Connectors

Rill is continually adding new sources and connectors in our releases. For a comprehensive list, you can refer to our [Connectors](/connect/source/) page. Please don't hesitate to [reach out](/contact) either if there's a connector you'd like us to add!

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
