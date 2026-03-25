<p align="center">
    <a href="https://rilldata.com/" target="_blank">
        <img width="10%" src="https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg" alt="Rill logo">
    </a>
</p>

<h3 align="center">Agent-first, human-friendly business intelligence</h3>

<p align="center">
  <a href="https://github.com/rilldata/rill/releases"><img src="https://img.shields.io/github/tag/rilldata/rill.svg" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/rilldata/rill.svg" alt="License"></a>
</p>

<p align="center">
  <a href="https://docs.rilldata.com/">Docs</a> · <a href="https://github.com/rilldata/rill-examples">Examples</a> · <a href="https://datatalks.rilldata.com/">Data Talks</a>
</p>

---

**Rill** is the fastest business intelligence tool for humans and agents, powered by embedded OLAP engines (ClickHouse, DuckDB and more!) and trusted by thousands of analysts and business users around the globe.

<p align="center">
  <img src="https://docs.rilldata.com/img/explore/dashboard101/multi-measure-select.png" alt="Rill dashboard with agentic analytics" width="80%">
</p>

## Why Rill?

### Agentic Authoring

Rill's BI-as-code foundation enables agentic authoring of projects — allowing developers to create and edit with tools like Claude Code and Cursor. Semantic layer definitions, dashboard configurations, visualizations, color themes, and security policies can all be built, tested, and deployed with coding agents and Git workflows.

### Conversational Analytics with a Context Layer for Trusted Accuracy

Rill's conversational BI goes beyond dashboards with a natural language interface that powers instant, visual, and verifiable insights. Rill plugs Claude and ChatGPT into its embedded semantic layer and real-time OLAP engine, enabling higher fidelity with conversation-fast performance.

### Agents Need Speed: Designed for Real-Time Performance

While most BI tools avoid issuing highly concurrent, low-latency queries, Rill embraces them — designed from the ground up to leverage fast OLAP engines. This enables differentiated performance for exploratory dashboards, pivot tables, and conversational analytics. Rill's 'Explain' feature, an agent-powered root cause analysis path, issues dozens of successive queries yet returns in seconds.

### Agents Require Context: The Semantic Layer Belongs in the Database

Semantic layers are critical for accuracy and trustworthiness of agentic analytics. A [recent study by Google](https://cloud.google.com/blog/products/business-intelligence/how-lookers-semantic-layer-enhances-gen-ai-trustworthiness) showed that using a semantic layer reduces data errors in generative AI queries by two-thirds.

In Rill, a semantic layer is defined in YAML with text descriptions and aggregate expressions — `revenue = sum(price * quantity)` — then represented as materialized views in the OLAP engine, providing performance benefits.

### Agents Simplify Embedded Analytics: Deep Customization at Low Cost

Agentic authoring paths are especially important for embedded analytics, which require complex customization and control. Teams manage the configuration of hundreds of customer-facing dashboards using only coding agents, saving thousands of hours of development time.

## Quickstart

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

Or explore a [live embedded dashboard](https://rill-embedding-example.netlify.app/).

## Key Capabilities

- [**Embedded OLAP**](https://docs.rilldata.com/build/olap-engines/) — Managed ClickHouse or DuckDB built in, or bring your own OLAP engine
- [**Semantic Layer**](https://docs.rilldata.com/build/metrics-view/) — Single source of truth for dimensions, measures, and time grains
- [**Explore & Canvas Dashboards**](https://docs.rilldata.com/build/dashboards) — Interactive dashboards, embeddable in your product
- [**Conversational BI**](https://docs.rilldata.com/explore/ai-chat) — Ask questions in natural language, get visual, verifiable insights
- [**MCP Server**](https://docs.rilldata.com/explore/mcp) — Connect Claude, ChatGPT, or any AI agent to your metrics
- [**Custom APIs & Embedding**](https://docs.rilldata.com/build/custom-apis/) — Expose metrics via REST or embed dashboards in your product
- [**Git-Native Deploy**](https://docs.rilldata.com/deploy/deploy-dashboard/) — Version control your dashboards, CI/CD your analytics
- [**Row Access Policies**](https://docs.rilldata.com/build/metrics-view/security) — Control data access per user, group, or attribute
- [**Alerts & Reports**](https://docs.rilldata.com/build/alerts) — Code-defined or UI-defined alerting on metric thresholds

## Example

A complete Rill project in two files:

**`models/auction_data.yaml`** — import and transform with SQL

```yaml
type: model
materialize: true
connector: duckdb

sql: |
  select * from read_parquet('gs://rilldata-public/auction_data.parquet')
    where pub_name like '%TV%'
```

**`metrics/auction_data_metrics.yaml`** — define metrics

```yaml
version: 1
type: metrics_view

display_name: Auction Data Metrics
model: auction_data
timeseries: __time

dimensions:
  - name: app_site_name
    column: app_site_name
  - name: app_site_domain
    column: app_site_domain

measures:
  - name: total_bid_requests
    expression: SUM(bid_request_cnt)
    description: Total number of bid requests
    format_preset: humanize
  - name: total_has_bid_floor
    expression: SUM(has_bid_floor_cnt)
    description: Count of entries where a bid floor was present
    format_preset: humanize
```

## Production Examples

| Example              | Description                                                            | Links                                                                                                                                                                                                                      |
| -------------------- | ---------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Programmatic Ads** | Bidstream data for pricing strategies and campaign performance         | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) · [Walkthrough](https://docs.rilldata.com/guides/openrtb-analytics) · [Demo](https://ui.rilldata.com/demo/rill-openrtb-prog-ads)       |
| **Cost Monitoring**  | Cloud infrastructure merged with customer data for efficiency analysis | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring) · [Walkthrough](https://docs.rilldata.com/guides/cost-monitoring-analytics) · [Demo](https://ui.rilldata.com/demo/rill-cost-monitoring) |
| **GitHub Analytics** | Codebase activity, contributor productivity, and commit patterns       | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics) · [Walkthrough](https://docs.rilldata.com/guides/github-analytics) · [Demo](https://ui.rilldata.com/demo/rill-github-analytics)        |

## Community

Have questions, need support, or just want to talk about Rill?

[![Discord](https://img.shields.io/badge/Discord-Join%20Chat-7289da?logo=discord&logoColor=white)](https://discord.gg/2ubRfjC7Rh) [![Twitter](https://img.shields.io/badge/Twitter-Follow-1da1f2?logo=twitter&logoColor=white)](https://twitter.com/RillData) [![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-181717?logo=github&logoColor=white)](https://github.com/rilldata/rill/discussions)

## Contributing

We welcome contributions! See our [Contributing Guide](https://docs.rilldata.com/home/contribute) to get started.
