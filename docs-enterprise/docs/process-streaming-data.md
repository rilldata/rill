---
title: "Process Streaming Data"
slug: "process-streaming-data"
excerpt: "Load real-time data to Rill via streaming services"
hidden: false
createdAt: "2022-06-02T19:51:38.148Z"
updatedAt: "2022-07-13T07:10:23.466Z"
---
# Getting Started

For real-time use cases, Druid supports ingestion with multiple messaging services - though Apache Kafka is most frequent. The Kafka indexing service enables the configuration of supervisors, which facilitate ingestion from Kafka by managing the creation and lifetime of Kafka indexing tasks. These indexing tasks read events using Kafka's own partition and offset mechanism and are therefore able to provide guarantees of exactly-once ingestion. The supervisor oversees the state of the indexing tasks to coordinate handoffs, manage failures, and ensure that the scalability and replication requirements are maintained.

For details on Kafka ingestion, proceed to the [Confluent/Apache Kafka page](https://enterprise.rilldata.com/docs/tutorial-kafka-ingestion).