---
title: Why Rill?
sidebar_label: Why Rill?
sidebar_position: 10
---

Data teams often struggle with a fragmented stack: while data lakes and warehouses are cost-effective for large volumes, they often lack the sub-second performance required for interactive, operational analytics. Engineering teams spend too much time managing ETL pipelines and answering ad-hoc requests, while business users are stuck with slow, static dashboards that don't allow for true exploration.

Rill bridges this gap by bringing fast, interactive analytics directly to your data lake or warehouse, with a developer-friendly workflow.

## What Makes Rill Different?

- **Fast, interactive dashboards on large datasets**: Rill powers sub-second queries and instant interactions, even on hundreds of millions of rows, by optimizing aggregation and pruning.
- **BI-as-Code**: Manage your entire analytics stack with code (SQL and YAML), bringing version control, CI/CD, and collaboration to your dashboards.
- **Works directly with your data lake**: [Connects directly](/developer/build/connectors) to your cloud data warehouse (BigQuery, Snowflake), lakehouse, or object storage (S3, GCS, Azure) without complex ETL.
- **Built-in OLAP**: An [integrated, in-memory OLAP engine](/developer/build/connectors/olap) handles data orchestration and query execution automatically for blazing speed.

## How It Works

![architecture](/img/concepts/architecture/architecture.svg)

Rill automatically turns your SQL data models into interactive, exploratory dashboards. It ingests data from your external sources into an embedded, high-performance OLAP engine. By defining metrics and dimensions upfront, Rill aggregates and optimizes the data, creating a responsive interface that allows users to slice and dice billions of records instantly without waiting for query processing.

This approach means engineering teams can:
1.   **Orchestrate data** out of cloud [data warehouses](/developer/build/connectors/data-source#data-warehouses) or [object stores](/developer/build/connectors#object-storage) into the fast serving layer.
2.   **Define metrics & dimensions** in [metrics view](/developer/build/metrics-view) to automatically aggregate and prune raw tables.
3.  **Deploy** your [project to Rill Cloud](/developer/deploy/deploy-dashboard) without managing separate BI servers or database infrastructure.

## Built for Operational Analytics

Operational intelligence provides decision-making capabilities that are complementary to traditional Business Intelligence (BI).

*   **Business Intelligence** drives complex decisions that occur daily or weekly, often on historical data (e.g., "How did we perform last quarter?").
*   **Operational Intelligence** fuels fast, frequent decisions on real-time and near-time data by hands-on operators (e.g., "Why is site latency spiking right now?" or "Which campaign is driving traffic this hour?").

Rill is built for this "fast" mode. It empowers product managers, operations teams, and engineers to explore data at the speed of thought, identifying trends and anomalies as they happen, without waiting for pre-computed reports.

## Key Benefits

*   **Performance**: Rill achieves performance that end users love by serving dashboards from in-memory data models, often reducing the data footprint by 10–100x compared to raw events.
*   **Developer Experience**: Define your entire analytics stack—from data models to dashboards—using code. This brings the benefits of version control, collaboration, and automation that software development teams rely on.
*   **Cost Effective**: Build analytics directly on your storage. Rill lowers costs because data can be transformed in place without duplication in a heavy enterprise warehouse.
*   **AI-Powered**: Rill integrates with AI tools to let you query your metrics views using natural language, ensuring responses are accurate and governed by your predefined measures and dimensions. See [AI Chat](/guide/ai/ai-chat) for more.

