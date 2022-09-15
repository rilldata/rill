---
title: "Tutorial: Kafka Ingestion"
slug: "tutorial-kafka-ingestion"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

To get comfortable with Druid and Streaming from Apache Kafka, we'll walk you through loading a sample data set. If you have yet configured an Apache Kafka instance that is accessible by RillData, please see the documentation on [Connecting Sources for Kafka](/connecting-with-kafka).

## Tutorial: Load an Apache Kafka Topic into Apache Druid

This tutorial goes through building the specification through the UI. If you have multiple similar specifications to create, it is recommended going to `Edit Spec` and copying/pasting that spec and then start the UI process.  This will save a lot of redo/boiler-plate setup that you have already established from previous ingestions.

1. **Click on `Druid Console`**

  This is a button in the upper right of RCC. A new tab will be created for you that displays the Druid console. You'll see see `Load Data`, `Ingestion`, and `Query` tabs.

2. **Click on the `Load Data`**

  If you've loaded data recently, you will see two buttons that give you the choice of "Start a new spec" or "Continue from previous spec".  If you see these buttons, click Start a new spec.

3. **Select "Apache Kafka"** and then Click on **Connect data ->**. 

  ![](https://images.contentful.com/ve6smfzbifwz/21J4LkoRGji9a49h3jI15J/e0a68b1bc5bf6ef5919db1fcf82628fd/1eb9d71-Ingest_Kafka_Start.png)

Using the **bootstrap server**, **topic**, and **credentials** from your Apache Kafka Cluster, enter them here. For **Confluent Cloud**, the bootstrap server will be provided from the Cluster dashboard. Credentials should be obtained through the API tools also provided. See the Connecting Sources for Kafka for additional details on establishing connection information.

| Property | Description | Example |
| --- | --- | --- |
| Bootstrap server |  | foo.us-east4.gcp.confluent.cloud:9092 |
| Topic | the topic to stream into RillData | rilldata-json-example |
| Consumer properties | The bootstrap server will be copied from above, but all other Kafka consumer properties that are needed to consume, should be added.  At a minimum the security information. | {  "bootstrap.server": "foo.us-east4.gcp.confluent.cloud:9092", "security.protocol" : "SASL_SSL", "sasl.jaas.config" : "org.apache.kafka.common.security.plain.PlainLoginModule   required username='{{KEY}}'   password='{{SECRET}}';", "sasl.mechanism": "PLAIN"} |
| Where should the data be sampled from | read from earliest offset in the topic or start with what is currently being written. | Start of stream |


4. This loads the example data and displays your data, giving you a chance to verify that the data is being parsed as you expect. In this example we are looking at [OpenSky Network](https://opensky-network.org/) data which was populated to a Kafka topic using an open-source Kafka connector for the OpenSky Network data.
![](https://images.contentful.com/ve6smfzbifwz/5HmpU9MlW4alyC9Bwot1Pi/42fb63a0fcf2fd436aa6e18bc9885975/9817c2b-Ingest_Kafka_Connect.png)

  The tabs along the top of the page:  `Start`, `Connect`, `Parse data`, ... `Publish` represent stages in the ingestion process. In this example you will move from stage to stage by clicking the `Next` button in the bottom right of the page. Each time you click the `Next` button, Druid will move to the next stage, making its best guess as to the appropriate parameters.The highlighted tab at the top will indicate the stage you have just moved to and if you want to re-execute that stage with different parameters, you can change the parameters in the form at the right and then click `Apply`. 

  In the next steps you will walk through these stages of the ingestion process by clicking the button in the bottom right (currently `Next: Parse date`), but you can also move back and forth among the steps by clicking on the tabs at the top.
 
5. **Click on `Next: Parse data`**

  The data loader parses the data based on its best guess about the type and displays a data preview. In this case the data is json and it chooses json, as shown by the `input format` field. 
![](https://images.contentful.com/ve6smfzbifwz/3CHtMof0fm22xHVt0qCCZD/c3b5cbe733140bb94d3589b68dbfa1d8/407824e-Ingest_Kafka_Parse.png)

  If the data was not json, you could change this and click apply. Click on the `input format` field to get a sense of the other choices, and feel free to click apply, but make sure JSON is selected and the display shows your data before you proceed to the next step.

6. **Click `Next: Parse time`** 

  This step analyzes the data to identify a time column, and moves that column to the far left, with the column name __time. You can see that it is coming from the column originally labeled 'time'.
![](https://images.contentful.com/ve6smfzbifwz/60JhtZ8pDF3zdRUhIV4dRt/bdd116fb14873bf0ee39036b1a508b0b/62a773b-Ingest_Kafka_ParseTime.png)

  In this example there are 2 columns with time data, one is when the time was pulled from open-sky, and one is when open-sky obtained the information. When multiple choices are provided, you need to pick what makes sense to your business use-case for this dataset.

  Druid does not have the means to use the event time associated with the record, if you need to leverage the event time, that will need to be added to the content of the message on the topic.

:::caution Timestamp Required
Druid requires that you specify a timestamp column and does optimizations based on this column. If your data does not have a timestamp column, you can select `Constant value`, or if your data has multiple timestamp columns, this is your opportunity to select a different one. If you specify a different time column than the default, you click `Apply` to apply your new setting.
:::

  In this example, the data loader determines that the `time` column is the only candidate to be used as a time column. That's our only time column in this data set, so we leave it set as is.

7. **Click `Next: Transform`** 

  In this step we have the opportunity to transform  one or more columns or add new columns.  See the Data Ingestion Tutorial for an example transformation.

  There is a rich SQL set of functions available, see [https://druid.apache.org/docs/0.21.0/misc/math-expr.html](https://druid.apache.org/docs/0.21.0/misc/math-expr.html).

8. **Click on `Next: Filter`**

  Here you have the opportunity to filter out rows. To see the syntax, click on `filter` link in the help, which will take you to [https://druid.apache.org/docs/0.21.0/querying/filters.html](https://druid.apache.org/docs/0.21.0/querying/filters.html). 

9. **Click on `Next: Configure Schema`**. 

  This takes you to a stage where you have the ability to do all of the following:
    + specify what dimensions and metrics are included in your dataset
    + aggregate measures to a coarser level of detail
    + create new measures that represent aggregations based on Sketches or HLL. For example you can create a measure that represents an approximation of a count or a unique count of a dimension. 

  Druid will make assumptions based on the data-types, use this to correct those datatypes. For example, velocity, altitude, latitude, and longitude are not topically metrics so we will convert them from metrics back to dimensions of type double.

  Select a column and then use the information to adjust.
![](https://images.contentful.com/ve6smfzbifwz/4utzFK0w9aSBY9Qdrbz30T/4ba04faf494d77885b9c698d214a5280/5cc523f-Ingest_Kafka_ConfigureSchema.png)
     
10. **Click `Next: Partition`** 

  Here in the partition panel you can choose an optimal way to segment your data across the druid cluster. You'll be segmenting based on your time dimension and you can segment at the same granularity as your time aggregation or a coarser granularity. For example, if you've chosen to aggregate (rollup/query granularity) your data to the hour, you can choose to physically segment it by hour, by day, by week or by month. For now leave it set to the default, `Hour`'.

![](https://images.contentful.com/ve6smfzbifwz/4BSc8yLT2I2fgjxi6WroYX/3a6486e63a426e625eb615ce72bd7e8c/443843b-Ingest_Kafka_Partition.png)

11. **Click `Next: tune`**

  This allows you to tune the ingestion.

![](https://images.contentful.com/ve6smfzbifwz/e3SpAZJxxyfUwZPBflITF/dc038b3b0932727e7107a474d236adc5/3baffb1-Ingest_Kafka_Tune.png)

12. **Click `Next: publish`**
  
  Here you can choose the name of your new dataset by filling in the `Datasource name` field.

![](https://images.contentful.com/ve6smfzbifwz/zyy5zboRqfyGeZaRUSdh3/ec8870b5474dcfa3b70c22f7352a7c30/8ff26a7-Ingest_Kafka_Publish.png)

13. **Click `Next: Edit Spec`**

  The json representation of the spec that you just created is displayed. You could edit this by hand (and you can also generate this by hand). Right now we want to go ahead and create our dataset based on the ingestion spec as is, so we won't make any changes.
![](https://images.contentful.com/ve6smfzbifwz/2CZ8nvDVcFOyJqKWdt9F3y/dfcd6a8dfe709e1ff9e512a016e7343f/488746f-Ingest_Kafka_EditSpec.png)

14. **Click `Submit`** 

  You are taken to a panel that shows that status of your job. It will first show the job as running, and then when it's done, the "Running" string will change to "Success". Once it shows success, you can query your data.
![](https://images.contentful.com/ve6smfzbifwz/4r0Ak2ZZllSAasx10qQk9f/2cf72078d2264fb10a334c4116144ee6/be06f3a-Ingest_Kafka_Ingestion.png)

15. **Click `Query` ** (rightmost tab at the top)

  This takes you to the Druid SQL console where you can use SQL to query your data.  For example

```sql
SELECT
  __time,
  barometricAltitude,
  callsign,
  "count",
  geometricAltitude,
  heading,
  id,
  latitude,
  longitude,
  onGround,
  originCountry,
  positionSource,
  specialPurpose,
  timePosition,
  velocity,
  verticalRate
FROM "rilldata-json"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' DAY
```

![](https://images.contentful.com/ve6smfzbifwz/7y9dGRnlfxZA2qnHeVghpH/fc635a24d9e9d42a2b13fbe43b8b5b9d/f2bc8db-Ingest_Kafka_Query.png)

From here you can query your data using Druid SQL. Note that by default `Smart query limit` is set to 100. If you want more than 100 rows, turn this toggle off and use the `limit` SQL expression to specify your own limit. A description of Druid's SQL language can be found here: [https://druid.apache.org/docs/0.21.0/querying/sql.html](https://druid.apache.org/docs/0.21.0/querying/sql.html)