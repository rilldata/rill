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
  <a href="https://docs.rilldata.com/">Docs</a> · <a href="https://datatalks.rilldata.com/">Data Talks</a>
</p>

---

<p align="center">
  <img src="https://docs.rilldata.com/img/explore/dashboard101/multi-measure-select.png" alt="Rill dashboard" width="80%">
</p>

**Rill** is the fastest BI tool for humans and agents, powered by OLAP engines like ClickHouse and DuckDB.

## Get Started

```bash
curl https://rill.sh | sh        # install
rill start my-project            # create a project and open the UI
```

### Scaffold a project with agent context

Use `rill init` to scaffold a project interactively:

```
➜ rill init
? Project name my-rill-project
? OLAP engine duckdb
? Agent instructions claude

Created a new Rill project at ~/my-rill-project
Added Claude instructions in .claude and .mcp.json

Success! Run the following command to start the project:

  rill start my-rill-project
```

## Why Rill?

- **Build with agents** — BI-as-code (YAML + SQL) means coding agents like Claude Code and Cursor can author projects, dashboards, and security policies end-to-end
- **Semantic layer** — Single source of truth for dimensions, measures, and time grains — defined in YAML, generating SQL at query time against your OLAP engine
- **Explore with agents** — Conversational BI lets business users query metrics in natural language; the [MCP server](https://docs.rilldata.com/explore/mcp) connects AI agents directly to your semantic layer
- **Real-time performance** — Sub-second queries on billions of rows via ClickHouse, DuckDB, and other OLAP engines
- **Embeddable** — Dashboards, APIs, and agent interfaces you can ship in your product

## Capabilities

### Rill Developer (local)

- [**Connectors**](https://docs.rilldata.com/build/connectors/) — S3, GCS, databases, and 20+ sources
- [**OLAP Engines**](https://docs.rilldata.com/build/olap-engines/) — Managed ClickHouse or DuckDB included, or connect an external engine (ClickHouse Cloud, Druid, Pinot, MotherDuck)
- [**SQL Models**](https://docs.rilldata.com/build/models/) — Transform raw data with SQL, join models together
- [**Data Profiling**](https://docs.rilldata.com/build/models) — Instant column stats and distributions
- [**Incremental Ingestion**](https://docs.rilldata.com/build/models/incremental-models) — Load only new data on each run to keep large datasets current without full refreshes
- [**Semantic Layer**](https://docs.rilldata.com/build/metrics-view/) — Dimensions, measures, and time grains in YAML
- [**Row Access Policies**](https://docs.rilldata.com/build/metrics-view/security) — Per-user, per-group data access control
- [**Local Dashboards**](https://docs.rilldata.com/build/dashboards) — Preview and explore dashboards locally

### Rill Cloud

- [**Deploy**](https://docs.rilldata.com/deploy/deploy-dashboard/) — Push to Rill Cloud with `rill deploy` or automate via Git-triggered CI/CD
- [**Explore & Canvas Dashboards**](https://docs.rilldata.com/build/dashboards) — Interactive dashboards, embeddable in your product
- [**Conversational BI**](https://docs.rilldata.com/explore/ai-chat) — Ask questions in natural language
- [**MCP Server**](https://docs.rilldata.com/explore/mcp) — Connect Claude, ChatGPT, or any AI agent to your metrics
- [**Custom APIs & Embedding**](https://docs.rilldata.com/build/custom-apis/) — Expose metrics via REST or embed dashboards
- [**Alerts & Reports**](https://docs.rilldata.com/build/alerts) — Threshold alerting, code-defined or UI-defined

## How It Works

Define everything in code — models, metrics, dashboards — and Rill handles the rest.

**1. Connect data** — `models/events.yaml`

```yaml
type: model
connector: duckdb
materialize: true

sql: |
  select * from read_parquet('gs://rilldata-public/auction_data.parquet')
```

**2. Define metrics** — `metrics/events_metrics.yaml`

```yaml
version: 1
type: metrics_view
model: events
timeseries: timestamp

dimensions:
  - name: country
    column: country
  - name: device
    column: device_type

measures:
  - name: total_events
    expression: count(*)
  - name: revenue
    expression: sum(price * quantity)
    description: Total revenue
```

**3. Create a dashboard** — `dashboards/events_explore.yaml`

```yaml
type: explore

display_name: "Events Dashboard"
metrics_view: events_metrics

dimensions: "*"
measures: "*"
```

**4. Deploy**

```bash
rill deploy                      # push to Rill Cloud
```

Your metrics view is immediately queryable on Rill Cloud — add YAML files to configure dashboards, alerts, and custom APIs.

## Production Examples

| Example              | Description                                         | Links                                                                                                                                            |
| -------------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Programmatic Ads** | Bidstream data for pricing and campaign performance | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) · [Demo](https://ui.rilldata.com/demo/rill-openrtb-prog-ads) |
| **Cost Monitoring**  | Cloud infra merged with customer data               | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring) · [Demo](https://ui.rilldata.com/demo/rill-cost-monitoring)   |
| **GitHub Analytics** | Contributor activity and commit patterns            | [GitHub](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics) · [Demo](https://ui.rilldata.com/demo/rill-github-analytics) |

Or explore a [live embedded dashboard](https://rill-embedding-example.netlify.app/).

## Community

[![Discord](https://img.shields.io/badge/Discord-Join%20Chat-7289da?logo=discord&logoColor=white)](https://discord.gg/2ubRfjC7Rh) [![Twitter](https://img.shields.io/badge/Twitter-Follow-1da1f2?logo=twitter&logoColor=white)](https://twitter.com/RillData) [![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-181717?logo=github&logoColor=white)](https://github.com/rilldata/rill/discussions)

## Contributing

We welcome contributions! See our [Contributing Guide](https://docs.rilldata.com/home/contribute) to get started.
