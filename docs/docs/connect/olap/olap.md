---
title: "Connect to a OLAP Engine"
description: Configure the OLAP engine used by Rill
sidebar_label: "OLAP Engines"
sidebar_position: 00
toc_max_heading_level: 3
className: connect-connect
---

import ConnectorIcon from '@site/src/components/ConnectorIcon';

Along with supporting ingestion from your various [sources](/connect/data-source), we allow a "live connection" to your own OLAP analytics engine. You can either set the default OLAP engine for your entire project, or set it for specific dashboards. 

1. Set the [default OLAP engine](/reference/project-files/rill-yaml#configuring-the-default-olap-engine) via the rill.yaml file.
2. Set the [OLAP engine](/reference/project-files/metrics-views) for a specific dashboard, on the metrics view

:::warning OLAP Engine Limitations
**Important Considerations:**
- **Use a test database** to avoid accidentally overwriting production data
- **Perform modeling upstream** since this is a live connection to your existing data
- **Incremental processing and related queries are not supported**
- **Feature availability may vary** between different OLAP engines
:::


## OLAP Engines

Rill supports the use of several different OLAP engines to power your dashboards, including:

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


:::note Additional OLAP Engines
Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, refer to our [OLAP Engines](/connect/olap) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](/contact)!
:::

## Multiple OLAP Engines in a Single Project

Rill supports the use of multiple OLAP engines in a single project with some limitations. For more detailed information, see our reference on [multiple OLAP engines](/connect/olap/multiple-olap). The basic use cases for multiple engines in a single project are:

1. Using Rill on top of already created and optimized tables from different OLAP sources.
2. Separating data based on size, as performance on different engines differs based on the size of the data.

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
