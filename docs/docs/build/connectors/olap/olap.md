---
title: "Bring Your Own  OLAP Engine (Live Connector)"
description: Configure the OLAP engine used by Rill
sidebar_label: "OLAP Engines"
sidebar_position: 0
toc_max_heading_level: 3
className: connect-connect
---

import ConnectorIcon from '@site/src/components/ConnectorIcon';

Rill supports connecting directly to your own OLAP engine via a "live connector". In this mode, no data is ingested into Rill, and all compute is pushed down to the OLAP engine. Use this mode if you've already handled all of your modeling upstream and want to use Rill as your visual application layer.

:::tip Models on Live Connectors

Rill also offers the ability to ingest and create tables directly from a [data source](/build/connectors/data-source) to your OLAP engine via the live connector, however you'll need to consider a few topics.

- **Use a test database** to avoid accidentally overwriting production data
- **Incremental processing and related queries are not supported**
- **Feature availability may vary** between different OLAP engines

:::



In order to connect Rill to your OLAP Engine:
1. Create the connector via the UI 
2. [Create the YAML](/reference/project-files/connectors#olap-engines) and set the [default OLAP engine](/reference/project-files/rill-yaml#configuring-the-default-olap-engine) via the rill.yaml file.

:::note `olap_connector` in rill.yaml
When setting the OLAP Engine via the UI, the `olap_connector` key will automatically update the rill.yaml.
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
    icon={<img src="/img/build/connectors/icons/Logo-DuckDB.svg" alt="DuckDB" />}
    content="Add extra parameters to Rill's embedded DuckDB or connect your own."
    link="/build/connectors/olap/duckdb"
    linkLabel="Learn more"
    referenceLink="duckdb"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Clickhouse.svg" alt="ClickHouse" />}
    content="High-performance columnar database for real-time analytics and data warehousing."
    link="/build/connectors/olap/clickhouse"
    linkLabel="Learn more"
    referenceLink="clickhouse"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Motherduck.svg" alt="MotherDuck" />}
    content="Cloud-native DuckDB service for scalable analytics and data processing."
    link="/build/connectors/olap/motherduck"
    linkLabel="Learn more"
    referenceLink="motherduck"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Druid.svg" alt="Druid" />}
    content="Real-time analytics database designed for high-performance OLAP queries."
    link="/build/connectors/olap/druid"
    linkLabel="Learn more"
    referenceLink="druid"
  />

  <ConnectorIcon
    icon={<img src="/img/build/connectors/icons/Logo-Pinot.svg" alt="Pinot" />}
    content="Distributed OLAP datastore for real-time analytics and business intelligence."
    link="/build/connectors/olap/pinot"
    linkLabel="Learn more"
    referenceLink="pinot"
  />
</div>


:::note Additional OLAP Engines
Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, refer to our [OLAP Engines](/build/connectors/olap) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](/contact)!
:::

## Multiple OLAP Engines in a Single Project

Rill supports the use of multiple OLAP engines in a single project with some limitations. For more detailed information, see our reference on [multiple OLAP engines](/build/connectors/olap/multiple-olap). The basic use cases for multiple engines in a single project are:

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


## What is OLAP?

OLAP (or Online Analytical Processing) is a computational approach designed to enable rapid, multidimensional analysis of large volumes of data. With OLAP, data is typically organized into cubes instead of traditional two-dimensional tables, which can facilitate complex queries and data analysis in a way that is significantly more efficient and user-friendly for analytical tasks. In particular, OLAP databases can be especially well suited for BI use cases that require deep, multidimensional analysis or real-time / user-facing analytics and applications. Additionally, many modern OLAP databases are optimized to ingest large volumes of data, execute low-latency queries with high throughput, and process billions of rows quickly with an emphasis on speed and efficiency in data retrieval. 

Unlike traditional relational databases or data warehouses that are optimized for transaction processing (with a focus on CRUD operations), OLAP databases are designed for query speed and complex analysis. Rather than storing data in a row-oriented manner, optimizing for transactional efficiency and operational queries, most OLAP databases are columnar and use pre-aggregated multidimensional cubes to speed up analytical queries. This allows a broad range of ad hoc queries and analysis to be performed without needing predefined schemas that are tailored to specific queries, and it's this flexibility that enables the highly interactive slice-and-dice exploration of data that powers Rill dashboards. This paradigm allows OLAP to be particularly well-suited for organizations and teams that want to dive deep into and understand their data to support decision-making processes, where speed and flexibility in the actual data analysis are important. 

:::info Want to see OLAP in action?

Check [here](https://www.rilldata.com/case-studies) to see examples of use cases that can be powered by OLAP.

:::

### External OLAP tables

Rill supports creating and powering dashboards using existing tables from alternative [OLAP engines](/build/connectors/olap) that have been configured in a particular project. These tables are not managed by Rill—hence, external—but allow users to bring in separate tables or datasets that might already exist in another preferred OLAP database of choice. This prevents the need to unnecessarily ingest this data into Rill, especially if the table is already optimized for use by this other OLAP engine, and allows Rill to connect to the data directly (and submit analytical queries).

<img src = '/img/build/connectors/external-tables/external-olap-db.png' class='rounded-gif' />
<br />

## Performance Tips

### Data Lifecycle Management 

One common way to decrease overall data size and improve query performance (by scanning less data) is to roll up your data to higher time grains historically. Typically, this means taking hourly data and rolling up to daily data when the additional level of granularity is no longer necessary for business needs. Databases like Apache Druid have these lifecycle tools built in, or reach out to Rill with questions.

A couple of considerations when rolling data from lower to higher time grains:

- Daily data loses time zone querying as everything is rolled up to a single time zone (usually UTC)
- Consider hashed compaction when going from hourly to daily to reduce data size even further
- Watch out for rolling up metrics. Some metrics should be summed—but others (like a bid floor or campaign budget) should stay unique and be rolled up as a max

### Dimension Stripping

Dimension stripping is another tool to reduce data size by removing high cardinality fields that are not required for analysis. While this can be done upfront in the dataset, another practice would be to drop these fields at certain intervals when they no longer add business value. Most frequently, we see a couple of decision points where these fields are dropped:

- After a day to first week, dropping user level details no longer needed for monitoring
- After a week to multiple weeks, dropping "double click" level details that aren't needed for reporting (e.g., the minor release number on an Operating System field)
- After a month to months, dropping fields no longer interesting for analysis

### Sampling & Datasketches

There are times when you may look at sampling data feeds to trade data accuracy for lower costs and faster query speeds. Sampling involves sending only a percentage of your data, then extrapolating the values to get an estimate. Rill does not recommend sampling your primary KPIs, any records that require a join, or are tied to revenue. This filtered data should be decided in random fashion to not skew or bias the results. Please note, tracking uniques is not recommended if you choose to sample.

If looking to track uniques, but with smaller datasets and significantly improved performance, you can load unique values (IP addresses, user IDs, URLs, etc.) with [datasketches](https://datasketches.apache.org). There are multiple types of datasketches supported depending on your engine. At a high level, datasketches use algorithms to approximate unique values. Common use cases for datasketches include count distincts (campaign reach, unique visitors) and quantiles (time spent, frequency). Check out the [Apache Datasketches](https://datasketches.apache.org/docs/Architecture/MajorSketchFamilies.html) site for more details on methodology and use cases.

### Lookups

While joins can kill the performance of [OLAP engines](/build/connectors/olap), lookups (key-value pairs) are common to reduce data size and improve query speeds. Lookups can be done during ingestion time (a static lookup to enrich the source data) or at query time (dynamic lookups).

**Static Lookups** 

Static lookups are lookups that are ingested at processing time. When a record is being processed, if a match is found between the record and lookup's key, the lookup's corresponding value at that moment in time is extracted and carbon-copied into Druid for the records it processed.

Static lookups are best suited for:

- Dimensions with values that require a historical record for how they have changed over time
- Values that are never expected to change (leverage Dynamic Lookups if the values are expected to change)
- Extremely large lookups (hundreds of thousands of records or >50MB lookup file) to improve query performance

Customers typically store lookup values in S3 or GCS, and the lookup file is then updated by customers as needed and consumed by ETL logic.

**Dynamic Lookups**

Since static lookups transform and store the data permanently, any changes to the mapping would require reprocessing the entire dataset to ensure consistency. To address the case when values in a lookup are expected to change with time, we developed dynamic lookups. Dynamic lookups, also known as Query Time Lookups, are lookups that are retrieved at query time, as opposed to being used at ingestion time.

Benefits of dynamic lookups include:

- Historical continuity for dimensions that change frequently without reprocessing the entire dataset
- Time savings, because there is no dataset reprocessing required to complete the update
- Dynamic lookups are kept separate from the dataset. Thus, any human errors introduced in the lookup do not impact the underlying dataset
- Ability for users to create new dimension tables from metadata associated with a dimension table. For example, account ownership can change during the course of a quarter. In such cases, a dynamic lookup can be updated on the fly to reflect the most current changes
