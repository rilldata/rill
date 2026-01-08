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
  <a href="https://docs.rilldata.com/">Docs</a> · <a href="#quickstart">Quickstart</a> · <a href="https://github.com/rilldata/rill-examples">Examples</a> · <a href="#capabilities">Capabilities</a> · <a href="https://docs.rilldata.com/home/contribute">Contributing</a> · <a href="https://datatalks.rilldata.com/">Data Talks</a>
</p>

---

**Rill** is an open-source BI-as-code tool with an embedded OLAP database. Define all of your components in YAML, query in milliseconds, deploy via Git.

- **Sub-second queries** — DuckDB/ClickHouse embedded or BYO OLAP Engine, data and compute co-located
- **Metrics layer** — Single source of truth for dimensions, measures, and time grains
- **AI-ready** — Built-in chat + MCP server for Claude, ChatGPT, and custom agents
- **Git-native** — Version control your dashboards, CI/CD your analytics

<p align="center">
  <img src="/docs/static/img/explore/dashboard101/multi-measure-select.png" alt="Rill dashboard" width="80%">
</p>

## Community

Have questions, need support, or just want to talk about Rill?

<p align="center">
  <a href="https://discord.gg/2ubRfjC7Rh"><img src="https://img.shields.io/badge/Discord-Join%20Chat-7289da?logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://twitter.com/medriscoll"><img src="https://img.shields.io/badge/Twitter-Follow-1da1f2?logo=twitter&logoColor=white" alt="Twitter"></a>
  <a href="https://github.com/rilldata/rill/discussions"><img src="https://img.shields.io/badge/GitHub-Discussions-181717?logo=github&logoColor=white" alt="GitHub Discussions"></a>
</p>

## Quickstart

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

## Capabilities

### Rill Developer (local, open-source)

| Feature                                                                            | Description                                       |
| ---------------------------------------------------------------------------------- | ------------------------------------------------- |
| **Process Data**                                                                   |                                                   |
| [Connectors](https://docs.rilldata.com/build/connectors/)                          | Connect to S3, GCS, databases, and 20+ sources    |
| [SQL Models](https://docs.rilldata.com/build/models/)                              | Transform raw data with SQL, join models together |
| [Data Profiling](https://docs.rilldata.com/build/models)                           | Instant column stats and distributions            |
| [Incremental Ingestion](https://docs.rilldata.com/build/models/incremental-models) | Partitioned, incremental data loading             |
| [Modeling Tests](https://docs.rilldata.com/build/models/data-quality-tests)        | Validate your data transformations                |
| **Define and Visualize Metrics**                                                   |                                                   |
| [Metrics Views](https://docs.rilldata.com/build/metrics-view/)                     | Define dimensions & measures in YAML              |
| [Local Dashboards](https://docs.rilldata.com/build/dashboards)                     | Preview and explore dashboards locally            |

### Rill Cloud

| Feature                                                                      | Description                                              |
| ---------------------------------------------------------------------------- | -------------------------------------------------------- |
| **Deploy & Manage**                                                          |                                                          |
| [Cloud Deploy](https://docs.rilldata.com/deploy/deploy-dashboard/)           | `rill deploy` to push dashboards to Rill Cloud           |
| [Row Access Policies](https://docs.rilldata.com/build/metrics-view/security) | Control data access per user, group, or attribute        |
| [Alerts](https://docs.rilldata.com/build/alerts)                             | Code-defined or UI-defined alerting on metric thresholds |
| **Visualize**                                                                |                                                          |
| [Explore Dashboards](https://docs.rilldata.com/build/dashboards/explore)     | Preview and explore dashboards locally                   |
| [Canvas Dashboards](https://docs.rilldata.com/build/dashboards/canvas)       | Drag-and-drop canvas for bespoke layouts                 |
| [Pivot Tables](https://docs.rilldata.com/explore/dashboard-101/pivot)        | Multi-dimensional analysis                               |
| [Time Comparisons](https://docs.rilldata.com/explore/time-series)            | Period-over-period analysis built-in                     |
| [Bookmarks](https://docs.rilldata.com/explore/bookmarks)                     | Save and share exploration states                        |
| **AI**                                                                       |                                                          |
| [AI Chat](https://docs.rilldata.com/explore/ai-chat)                         | Ask questions in natural language, get insights          |
| [MCP Server](https://docs.rilldata.com/explore/mcp)                          | Connect Claude, ChatGPT, or any AI agent                 |
| **Integrate**                                                                |                                                          |
| [Custom APIs](https://docs.rilldata.com/build/custom-apis/)                  | Expose metrics via REST endpoints                        |
| [Public URLs](https://docs.rilldata.com/explore/public-url)                  | Share dashboards without authentication                  |
| [Embedding](https://docs.rilldata.com/integrate/embedding)                   | Embed dashboards in your product                         |
| [Themes](https://docs.rilldata.com/build/dashboards/customization)           | Customize colors and branding                            |

→ [Try AI Chat live](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/-/ai) · [Embedding demo](https://rill-embedding-example.netlify.app/)

## Example

A complete Rill project in four files:

**`connectors/s3.yaml`** — connect to data

```yaml
# Connector YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/connectors

type: connector

driver: gcs
google_application_credentials: "{{ .env.connector.gcs_1.google_application_credentials }}"
```

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
# Metrics view YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

display_name: Auction Data Metrics
model: auction_data
timeseries: __time

dimensions:
  - name: app_site_name
    display_name: App Site Name
    column: app_site_name
  - name: app_site_domain
    display_name: App Site Domain
    column: app_site_domain

measures:
  - name: total_bid_requests_measure
    display_name: Total Bid Requests
    expression: SUM(bid_request_cnt)
    description: Total number of bid requests recorded in the table.
    format_preset: humanize
  - name: total_has_bid_floor_measure
    display_name: Total Has Bid Floor Count
    expression: SUM(has_bid_floor_cnt)
    description: Total count of entries where a bid floor was present.
    format_preset: humanize
```

**`dashboards/auction_data_explore.yaml`** — define visualization

```yaml
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

display_name: "Auction Data Metrics dashboard"
metrics_view: auction_data_metrics

dimensions: "*"
measures: "*"
```

→ [More examples](https://github.com/rilldata/rill-examples) | [Live demos](https://ui.rilldata.com/demo)

## Production Examples

| Example              | Description                                                                 | Links                                                                                                                                                                                                                      |
| -------------------- | --------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Programmatic Ads** | Bidstream data for optimizing pricing strategies and campaign performance   | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) · [Walkthrough](https://docs.rilldata.com/guides/openrtb-analytics) · [Demo](https://ui.rilldata.com/demo/rill-openrtb-prog-ads)       |
| **Cost Monitoring**  | Cloud infrastructure data merged with customer data for efficiency analysis | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring) · [Walkthrough](https://docs.rilldata.com/guides/cost-monitoring-analytics) · [Demo](https://ui.rilldata.com/demo/rill-cost-monitoring) |
| **GitHub Analytics** | Analyze codebase activity, contributor productivity, and commit patterns    | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics) · [Walkthrough](https://docs.rilldata.com/guides/github-analytics) · [Demo](https://ui.rilldata.com/demo/rill-github-analytics)        |
| **App Engagement**   | Conversion dataset for analyzing funnel steps                               | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement) · [Demo](https://ui.rilldata.com/demo/rill-app-engagement)                                                                               |
| **Kitchen Sink**     | Deep dive into Rill features using ClickHouse GitHub data                   | [GitHub](https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial) · [Walkthrough](https://docs.rilldata.com/guides/rill-basics/launch) · [Demo](https://ui.rilldata.com/demo/my-rill-tutorial)                |

## Playlists

<div align="center">

[Getting Started with Rill](https://www.youtube.com/watch?v=oQSok8Dy-D0) · [Exploring Data](https://www.youtube.com/watch?v=wTP46eOzoCk&list=PL_ZoDsg2yFKgi7ud_fOOD33AH8ONWQS7I&index=1) · [Data Talks](https://www.youtube.com/playlist?list=PL_ZoDsg2yFKgr_YEc4XOY0wlRLqzyR07q)

</div>

## Contributing

We welcome contributions! See our [Contributing Guide](https://docs.rilldata.com/home/contribute) to get started.

## License

[Apache 2.0](LICENSE)
