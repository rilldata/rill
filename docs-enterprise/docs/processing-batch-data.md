---
title: "Process Batch Data"
slug: "processing-batch-data"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Load data into Rill's data store via storage locations" />

## Getting Started
Many customers start with Rill loading data from storage locations - s3, GCS, etc. 

Data is loaded into Rill once client processing is complete (or potentially raw data) at some regular interval (usually hourly). 

## Load your own data

If your data is already in aggregate or final format, you can load directly into Rill: 
  * test your data manually ([more details here](https://druid.apache.org/docs/latest/ingestion/index.html)) which will create your Druid spec for ingestion
  *  using that ingestion spec, add an orchestration step post-processing to load data into Druid. any scheduling tool will work - we typically use Ariflow. [See the this example Airflow dag](https://github.com/gorillio/airflow-druid-examples) for more details
  * if you plan to use Rill Explore, contact support@rilldata.com to create your first [staging dashboard](/getting-started) for review

:::info Batch Ingestion
For more details on Druid ingestion, visit: 
  * [Ingestion Spec](https://druid.apache.org/docs/latest/ingestion/ingestion-spec.html) 
  * [Schema Design](https://druid.apache.org/docs/latest/ingestion/schema-design.html) 
  * [Native Batch Ingestion](https://druid.apache.org/docs/latest/ingestion/native-batch.html) 
:::


## Rill managed pipelines

For customers with more complex joins/transformations requiring Rill managed pipelines: 
  * grant Rill access to the storage location (usually [Amazon s3](/aws-s3-bucket) or [Google Cloud Storage](/gcs-bucket))
  * Rill to develop pipeline logic as required
  * review the sample output with the Rill team to confirm layout and values
  * Rill to poll source locations on regular intervals

:::caution Rill Managed Pipelines
Email [contact@rilldata.com](mailto:contact@rilldata.com) if you're interested in having the Rill team build and manage ingestion for your data pipelines
:::