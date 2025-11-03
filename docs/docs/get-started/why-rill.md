---
title: Why Use Rill?
sidebar_label: Why Use Rill? 
sidebar_position: 00
---

Rill's strategy for fast dashboards is threefold:

1) *Define metrics & dimensions up front*, and use these definitions to automatically aggregate and prune the raw tables. This modest modeling pain yields massive gains: the data footprint of aggregated metrics is typically 10–100x smaller than the underlying raw events in data lakes or warehouses.

2) *Use an integrated OLAP database* to drive dashboards, by orchestrating (and aggregating, per above) data out of cloud data warehouses, lakehouses, or object stores.

3) *Implement BI-as-code* to bring version control, collaboration, and automation to your analytics stack. This approach allows engineering teams to maintain control while enabling business users to make UI-based adjustments when needed. [Learn more about BI-as-code](#bi-as-code).

The decoupling of BI applications and database servers served a purpose at one phase in the evolution of data stacks, but the cost and performance trade-offs have begun to shift in favor of consolidated analytics offerings.

## Architecture

![architecture](/img/concepts/architecture/architecture.svg)


Engineering teams frequently get ad hoc requests from product, finance, and operations for insights into data sets that are readily available in object storage. Over time, writing complex SQL or Python reports against the data lake can become a burden and a distraction. With Rill, engineering teams can write SQL to design and deploy a small set of data models to answer a broad range of inquiries. Rill's architecture automatically turns SQL-based data models into interactive, exploratory dashboards with no additional design effort. Engineers can focus on defining expressions for core business metrics in SQL, and Rill takes care of the rest.

Building analytics directly on data storage reduces complexity because there are fewer moving pieces to manage, and it also lowers costs because data can be transformed in place without duplication. Rill achieves performance that end users love by serving dashboards from in-memory data models fed by data lake sources.

In short, Rill helps engineering teams leverage the value of their data lake architectures, delighting business users without requiring multiple investments in separate ETL, data warehouse, and BI tools.


## Operational vs. Traditional BI

The distinction between operational and business intelligence is analogous to the distinction between fast and slow thinking, as characterized by the psychologist Daniel Kahneman in his book __Thinking, Fast and Slow__. One system operates quickly and automatically for simple decisions, and the other leverages slow and effortful deliberation for complex decisions. 

Ultimately, the output of both operational and business intelligence is decisions. Operational intelligence fuels fast, frequent decisions on real-time and near-time data by hands-on operators. Business intelligence drives complex decisions that occur daily or weekly, on fairly complete data sets. 


<img src = '/img/concepts/operational/comparison.png' class='rounded-gif' />
<br />

### Why Operational BI requires new tools

Operational intelligence provides a set of decision-making capabilities that are complementary to business intelligence, but its unique performance requirements also demand a novel stack of distinct technologies which are complementary and sit adjacent to existing business intelligence stacks.

Analytics technology stacks can be thought of as data flowing into a three-layered cake consisting of ETL, databases, and applications. The requirements for an operational intelligence stack are that it supports:

- high speed of data from ETL to application
- high frequency, low-latency queries between the database and application layer

In the diagram below we illustrate two common examples for technologies used in operational and business intelligence stacks.

![operationalbi](/img/concepts/operational/operational.png)



## BI-As-Code 

BI-as-code is a modern approach to business intelligence that treats analytics assets as code, bringing the same benefits of version control, collaboration, and automation that software development teams have enjoyed for years. With Rill, you can define your entire analytics stack—from data models to dashboards—using code, while still maintaining the flexibility to make UI-based adjustments when needed.

<div style={{ textAlign: 'center' }}>
  <img src="/img/concepts/metrics-view/metrics-view-components.png" style={{ width: '100%', borderRadius: '15px', padding: '20px' }} />
</div>


This means that Rill projects can be completely generated via code agents that can be provided context on your specific use case and you can integrate Rill into your workflow to create and/or update your visualization via natural language.

### AI-Powered Analytics Development

Rill's BI-as-code approach enables powerful AI integration that transforms how analytics are created and maintained:

**Code Agent Integration**
- **Automated Project Generation**: AI agents can analyze your data schema and business requirements to automatically generate complete Rill projects
- **Context-Aware Development**: Provide your AI agent with specific use cases, data sources, and business metrics to create tailored analytics solutions
- **Intelligent Code Generation**: Agents can write SQL models, define metrics views, and configure dashboards based on your data structure and analytical needs

**Natural Language Workflows**
- **Conversational Analytics**: Create and modify dashboards through natural language commands like "Show me revenue trends by region" or "Modify the defaule view to show last quarter data"
- **Iterative Refinement**: Use natural language to refine existing visualizations: "Make this dashboard show monthly data instead of daily" or "Add a comparison with last year"
- **Business User Friendly**: Non-technical team members can request changes and new analyses without writing code or learning complex interfaces


This AI-powered approach democratizes analytics development while maintaining the rigor and control that engineering teams require, enabling faster iteration and broader participation in data-driven decision making. 