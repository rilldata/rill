---
title: "Process Batch Data"
slug: "processing-batch-data"
excerpt: "Load data into Rill's data store via storage locations"
hidden: false
createdAt: "2022-06-02T19:48:37.780Z"
updatedAt: "2022-07-13T07:07:29.774Z"
---
# Getting Started
Many customers start with Rill loading data from storage locations - s3, GCS, etc. 

Data is loaded into Rill once client processing is complete (or potentially raw data) at some regular interval (usually hourly). 

#Load your own data

If your data is already in aggregate or final format, you can load directly into Rill: 
  * test your data manually ([more details here](https://druid.apache.org/docs/latest/ingestion/index.html)) which will create your Druid spec for ingestion
  *  using that ingestion spec, add an orchestration step post-processing to load data into Druid. any scheduling tool will work - we typically use Ariflow. [See the this example Airflow dag](https://github.com/gorillio/airflow-druid-examples) for more details
  * if you plan to use Rill Explore, contact support@rilldata.com to create your first [staging dashboard](https://enterprise.rilldata.com/docs/getting-started) for review
[block:callout]
{
  "type": "info",
  "title": "Batch Ingestion",
  "body": "For more details on Druid ingestion, visit: \n  * [Ingestion Spec](https://druid.apache.org/docs/latest/ingestion/ingestion-spec.html) \n  * [Schema Design](https://druid.apache.org/docs/latest/ingestion/schema-design.html) \n  * [Native Batch Ingestion](https://druid.apache.org/docs/latest/ingestion/native-batch.html) "
}
[/block]
#Rill managed pipelines

For customers with more complex joins/transformations requiring Rill managed pipelines, 
  * grant Rill access to the storage location (usually [Amazon s3](https://enterprise.rilldata.com/docs/aws-s3-bucket) or [Google Cloud Storage](https://enterprise.rilldata.com/docs/gcs-bucket)
  * Rill to develop pipeline logic as required
  * review the sample output with the Rill team to confirm layout and values
  * Rill to poll source locations on regular intervals
[block:callout]
{
  "type": "warning",
  "title": "Rill Managed Pipelines",
  "body": "Email [contact@rilldata.com](mailto:contact@rilldata.com) if you're interested in having the Rill team build and manage ingestion for your data pipelines"
}
[/block]