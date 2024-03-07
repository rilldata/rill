---
title: "Performance Optimization"
description: Performance Optimization
sidebar_label: "Perfomance Optimization"
sidebar_position: 20
---

## Overview

## Query Optimization

## OLAP Engines

Often, data in external OLAP engines is quite large in size and requires different considerations to improve performance. At Rill, we manage clusters in the 100's of TB so included some tips below based on our experience.

### Data Lifecycle Management 

One common way to decrease overall data size and improve query performance (by scanning less data) is to rollup your data to higher time grains historically. Typically, this means taking hourly data and rollup up to daily data when the additional level of granularity is no longer necessary for business needs. Databases like Apache Druid have these lifecycle tools built in or reach out to Rill with questions.

A couple of considerations when rolling data from lower to higher time grains:

- Daily data loses time zone querying as everything is rolled up to a single time zone (usually UTC)
- Consider hashed compaction when going from hourly to daily to reduce data size even further
- Watch out for rolling up metrics. Some metrics should be summed - but others (like a bid floor or campaign budget) should stay unique and be rolled up as a max

### Dimension Stripping

Dimension stripping is another tool to reduce data size by removing high cardinality fields that are not required for analysis. While this can be done upfront in the data set, another practice would be to drop these fields at certain intervals when they no longer add busiess value. Most frequently, we see a couple decision points where these fields are dropped:

- After a day to first week, dropping user level details no longer needed for monitoring
- After a week to multiple weeks, dropping "double click" level details that aren't needed for reporting (e.g. the minor release number on an Operating System field)
- After a month to months, dropping fields no longer interesting for analysis

### Sampling & Datasketches

There are times where you may look at sampling data feeds to trade data accuracy for lower costs and faster query speeds. Sampling involves sending only a percentage of your data, then extrapolating the values to get an estimate. Rill does not recommend sampling your primary KPIs, any records that require a join or are tied to revenue. This filtered data should be decided in random fashion to not skew or bias the results. Please note, tracking uniques is not recommended if you choose to sample. 

If looking to track uniques, but with smaller datasets and significantly improved performance, you can load unique values (ip addresses, user ids, URLs, etc) with [datasketches](https://datasketches.apache.org/). There are multiple types of datasketches supported depending on your engine. At a high level, datasketches use algorithms to approximate unique values. Common use cases for datasketches include count distincts (campaign reach, unique visitors) and quantiles (time spent, frequency). Check out the [Apache Datasketches](https://datasketches.apache.org/docs/Architecture/MajorSketchFamilies.html) site for more details on methodology and use cases.

### Lookups

While joins can kill the performance of OLAP engines, Lookups (key-value pairs) are common to reduce data size and improve query speeds. Lookups can be done during ingestion time (a static lookup to enrich the source data) or at query time (dynamic lookups).

**Static Lookups** 
Static Lookups are lookups that are ingested at processing time. When a record is being processed, if a match is found between the record and lookup's key, the lookup's corresponding value at that moment in time is extracted and carbon-copied into Druid for the records it processed.

Static lookups are best suited for:

- Dimensions with values that require a historical record for how it has changed over time
- Values are never expected to change (leverage Dynamic Lookups if the values are expected to change)
- Extremely large lookups (hundreds of thousands or records or >50MB lookup file) to improve query performance

Customers typically store lookup values in s3 or GCS and the lookup file is then updated by customers as needed and consumed by ETL logic.

**Dynamic Lookups**

Since static lookups transform and store the data permanently, any changes to the mapping would require reprocessing the entire data set to ensure consistency. To address the case when values in a lookup are expected to change with time we, developed dynamic lookups. Dynamic Lookups, also known as Query Time Lookups, are lookups that retrieved at query time, as apposed to being used at ingestion time.

Benefits of dynamic lookups include:

- Historical continuity for dimensions that change frequently without reprocessing the entire data set
- Time savings, because there is no data set reprocessing required to complete the update
- Dynamic lookups are kept separate from the data set. Thus, any human errors introduced in the lookup do not impact the underlying data set
- Ability for users to create new dimension tables from metadata associated with a dimension table. For example, account ownerships can change during the course of a quarter. In such cases, a dynamic lookup can be updated on the fly to reflect the most current changes