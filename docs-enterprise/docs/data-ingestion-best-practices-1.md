---
title: "Ingestion Best Practices"
slug: "data-ingestion-best-practices-1"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

:::info Apache Druid Ingestion Details
For a full list of ingestion best practices, visit the [Apache Druid Ingestion guide](https://druid.apache.org/docs/latest/ingestion/index.html).
:::

## Data Clean-up

To get the most out of Rill, we suggest any fields sent are sent are "clean" for two reasons:

First, having a concise, low cardinality field will give your end users an easier time in finding the data they need. Having three values such as "iphone", "iPhone", and "Apple IPhone" will require the user to know the difference and select altogether when trying to view a single cohort. 

Secondly, dashboard & query performance is directly affected by the cardinality of fields. Replacing nonsensical values with a default "Not Available" can make a significant difference.
## Data Sampling

There are times where you may look at sampling data feeds to trade data accuracy for lower costs and faster query speeds. Sampling involves sending only a percentage of your data, then extrapolating the values to get an estimate. This filtered data should be decided in random fashion to not skew or bias the results. Please note, tracking uniques is not recommended if you choose to sample.

If you choose to sample, you should thereafter treat metrics as an indicative of trends, and not as a full representative of the truth. In particular, Rill does not recommend sampling your primary KPIs, any records that require a join or are tied to revenue.
## Lookup Tables

Based on the how often the raw data's supporting "lookup table" values update with time and whether or not your users would like to keep a historic view of how a dimension has changed over time, Rill will work with your team to apply one of our various lookup table solutions.

### Static Lookups
Static Lookups, also otherwise known as Ingestion Time Lookups, are lookups that are ingested at processing time. When a record is being processed, if a match is found between the record and lookup's key, the lookup's corresponding value at that moment in time is extracted and carbon-copied into Druid for the records it processed.

Static lookups are best suited for:
  * Dimensions with values that require a historical record for how it has changed over time
  * Values are never expected to change (leverage Dynamic Lookups if the values are expected to change)
  * Extremely large lookups (hundreds of thousands or records or >50MB lookup file) to improve query performance 

Customers typically store lookup values in s3 or GCS and the lookup file is then updated by customers as needed and consumed by ETL logic.

### Dynamic Lookups
Since static lookups transform and store the data permanently, any changes to the mapping would require reprocessing the entire data set to ensure consistency.

To address the case when values in a lookup are expected to change with time we, developed dynamic lookups. Dynamic Lookups, also known as Query Time Lookups, are lookups that retrieved at query time, as apposed to being used at ingestion time.

Benefits of dynamic lookups include:
  * Historical continuity for dimensions that change frequently without reprocessing the entire data set
  * Time savings, because there is no data set reprocessing required to complete the update
  * Dynamic lookups are kept separate from the data set. Thus, any human errors introduced in the lookup do not impact the underlying data set
  * Ability for users to create new dimension tables from metadata associated with a dimension table. For example, account ownerships can change during the course of a quarter. In such cases, a dynamic lookup can be updated on the fly to reflect the most current changes

Dynamic lookups are typically stored in s3 or GCS and then replicated into a Druid lookup table at a regular polling frequency (e.g. hourly, daily).
:::caution 1:1 (or 1:Many) vs Many:Many Lookups
Lookups with a unique key (1:1 or 1:Many) can be improved with specifying "injective" : true.

Many:Many lookups do not have this capability. As such, for large many:many lookups, we recommend considering lookup during processing / ETL.

For more info on lookups, [review Druid documentation here](https://druid.apache.org/docs/latest/querying/lookups.html).