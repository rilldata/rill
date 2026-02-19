<p align="center">
    <a href="https://rilldata.com/" target="_blank">
        <img width="10%" src="https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg" alt="Rill logo">
    </a>
</p>

<h3 align="center">The fastest path from data to dashboard</h3>

<p align="center">
  <a href="https://github.com/rilldata/rill/releases"><img src="https://img.shields.io/github/tag/rilldata/rill.svg" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/rilldata/rill.svg" alt="License"></a>
</p>

<p align="center">
  <a href="https://docs.rilldata.com/">Docs</a> · <a href="#quickstart">Quickstart</a> · <a href="https://github.com/rilldata/rill-examples">Examples</a> · <a href="#capabilities">Capabilities</a> · <a href="https://datatalks.rilldata.com/">Data Talks</a>
</p>

---

**Rill** is an open-source BI-as-code tool with an embedded OLAP database. Define all of your components in YAML, query in milliseconds, deploy via Git.

- [**Sub-second queries**](https://docs.rilldata.com/build/connectors/) — DuckDB/ClickHouse embedded or BYO OLAP Engine, data and compute co-located
- [**Metrics layer**](https://docs.rilldata.com/build/metrics-view/) — Single source of truth for dimensions, measures, and time grains
- [**AI-ready**](https://docs.rilldata.com/explore/ai-chat) — Built-in chat + MCP server for Claude, ChatGPT, and custom agents
- [**Git-native**](https://docs.rilldata.com/deploy/deploy-dashboard/) — Version control your dashboards, CI/CD your analytics

<p align="center">
  <img src="https://docs.rilldata.com/img/explore/dashboard101/multi-measure-select.png" alt="Rill dashboard" width="80%">
</p>


## Why Rill?

- [**Full-stack analytics**](https://docs.rilldata.com/) — OLAP + modeling + metrics + visualization in one deployment
- [**Code + AI**](https://docs.rilldata.com/explore/ai-chat) — Engineers get code-first; business users get AI-first; both are first-class
- [**Fast at every layer**](https://docs.rilldata.com/build/models/) — Sub-second queries on large datasets, hours from raw data to production
- [**Low barrier to entry**](#quickstart) — Two commands to start, free tier, managed cloud, or self-hosted

## Quickstart

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

> [!TIP]
> **Try it live:** [AI Chat demo](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/-/ai) · [Embedding demo](https://rill-embedding-example.netlify.app/) · [All live demos](https://ui.rilldata.com/demo)

## Capabilities

### Rill Developer (local, open-source)

- [**Connectors**](https://docs.rilldata.com/build/connectors/) — Connect to S3, GCS, databases, and 20+ sources
- [**SQL Models**](https://docs.rilldata.com/build/models/) — Transform raw data with SQL, join models together
- [**Data Profiling**](https://docs.rilldata.com/build/models) — Instant column stats and distributions
- [**Incremental Ingestion**](https://docs.rilldata.com/build/models/incremental-models) — Partitioned, incremental data loading
- [**Modeling Tests**](https://docs.rilldata.com/build/models/data-quality-tests) — Validate your data transformations
- [**Metrics Views**](https://docs.rilldata.com/build/metrics-view/) — Define dimensions & measures in YAML
- [**Local Dashboards**](https://docs.rilldata.com/build/dashboards) — Preview and explore dashboards locally

### Rill Cloud

- [**Cloud Deploy**](https://docs.rilldata.com/deploy/deploy-dashboard/) — `rill deploy` to push dashboards to Rill Cloud
- [**Row Access Policies**](https://docs.rilldata.com/build/metrics-view/security) — Control data access per user, group, or attribute
- [**Alerts**](https://docs.rilldata.com/build/alerts) — Code-defined or UI-defined alerting on metric thresholds
- [**Explore Dashboards**](https://docs.rilldata.com/build/dashboards/explore) — Interactive dashboards for slicing and dicing metrics
- [**Canvas Dashboards**](https://docs.rilldata.com/build/dashboards/canvas) — Drag-and-drop canvas for bespoke layouts
- [**AI Chat**](https://docs.rilldata.com/explore/ai-chat) — Ask questions in natural language, get insights
- [**MCP Server**](https://docs.rilldata.com/explore/mcp) — Connect Claude, ChatGPT, or any AI agent
- [**Custom APIs & Embedding**](https://docs.rilldata.com/build/custom-apis/) — Expose metrics via REST or embed dashboards in your product

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

> [!TIP]
> [More examples](https://github.com/rilldata/rill-examples) · [Live demos](https://ui.rilldata.com/demo)

## Production Examples

| Example | Description | Links |
| --- | --- | --- |
| **Programmatic Ads** | Bidstream data for pricing strategies and campaign performance | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) · [Walkthrough](https://docs.rilldata.com/guides/openrtb-analytics) · [Demo](https://ui.rilldata.com/demo/rill-openrtb-prog-ads) |
| **Cost Monitoring** | Cloud infrastructure merged with customer data for efficiency analysis | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring) · [Walkthrough](https://docs.rilldata.com/guides/cost-monitoring-analytics) · [Demo](https://ui.rilldata.com/demo/rill-cost-monitoring) |
| **GitHub Analytics** | Codebase activity, contributor productivity, and commit patterns | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics) · [Walkthrough](https://docs.rilldata.com/guides/github-analytics) · [Demo](https://ui.rilldata.com/demo/rill-github-analytics) |

## Community

Have questions, need support, or just want to talk about Rill?

[![Discord](https://img.shields.io/badge/Discord-Join%20Chat-7289da?logo=discord&logoColor=white)](https://discord.gg/2ubRfjC7Rh) [![Twitter](https://img.shields.io/badge/Twitter-Follow-1da1f2?logo=twitter&logoColor=white)](https://twitter.com/RillData) [![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-181717?logo=github&logoColor=white)](https://github.com/rilldata/rill/discussions)

## Contributing

We welcome contributions! See our [Contributing Guide](https://docs.rilldata.com/home/contribute) to get started.

## License

[Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0)
