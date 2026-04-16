---
title: Kafka
description: Stream data from Kafka into ClickHouse
sidebar_label: Kafka
sidebar_position: 75
---

import WrongOLAP from '@site/src/components/WrongOLAP';
import ClickHousePrereq from '@site/src/components/ClickHousePrereq';

<WrongOLAP engine="clickhouse" solo />

## Overview

[Apache Kafka](https://kafka.apache.org/) is a distributed event streaming platform used for building real-time data pipelines. ClickHouse can consume data from Kafka topics using its [Kafka table engine](https://clickhouse.com/docs/en/engines/table-engines/integrations/kafka), enabling near-real-time analytics on streaming data.

<ClickHousePrereq />

## How It Works

Unlike other ClickHouse data sources that use table functions in a model's SQL, Kafka integration is configured directly in ClickHouse using the Kafka table engine. Data flows through three components:

1. **Kafka engine table** — connects to the Kafka topic and consumes messages
2. **Materialized view** — transforms and routes data from the Kafka table
3. **Target MergeTree table** — stores the data for querying

Once your ClickHouse cluster has this pipeline set up, you can query the target table directly in your Rill models.

## Model Configuration

After setting up the Kafka pipeline in ClickHouse, create a model that queries the target table:

```yaml
type: model
connector: my_clickhouse

sql: |
  SELECT *
  FROM kafka_target_table
```

## Setting Up the Kafka Pipeline in ClickHouse

Below is an example of the ClickHouse-side setup. Run these statements directly in your ClickHouse cluster.

### 1. Create the target table

```sql
CREATE TABLE kafka_target_table (
  event_id String,
  event_type String,
  payload String,
  created_at DateTime
) ENGINE = MergeTree()
ORDER BY (event_type, created_at);
```

### 2. Create the Kafka engine table

```sql
CREATE TABLE kafka_source (
  event_id String,
  event_type String,
  payload String,
  created_at DateTime
) ENGINE = Kafka()
SETTINGS
  kafka_broker_list = 'broker1:9092,broker2:9092',
  kafka_topic_list = 'my-topic',
  kafka_group_name = 'clickhouse-consumer-group',
  kafka_format = 'JSONEachRow';
```

### 3. Create the materialized view

```sql
CREATE MATERIALIZED VIEW kafka_consumer TO kafka_target_table AS
SELECT *
FROM kafka_source;
```

:::tip
For Confluent Cloud or other secured Kafka clusters, add SASL/SSL settings to the Kafka engine table. See the [ClickHouse Kafka engine documentation](https://clickhouse.com/docs/en/engines/table-engines/integrations/kafka) for all available settings.
:::

## Reference

See the [ClickHouse Kafka engine documentation](https://clickhouse.com/docs/en/engines/table-engines/integrations/kafka) for full configuration options, including consumer group management, error handling, and supported data formats.
