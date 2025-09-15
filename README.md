<p align="center">
    <a href="https://rilldata.com/" target="_blank">
        <img width="10%" src="https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg" alt="Rill logo">
    </a>
</p>

<br/>
<p align="center">
    <a href="LICENSE" target="_blank">
        <img src="https://img.shields.io/github/license/rilldata/rill.svg" alt="GitHub license">
    </a>
    <a href="https://github.com/rilldata/rill/releases" target="_blank">
        <img src="https://img.shields.io/github/tag/rilldata/rill.svg" alt="GitHub tag (latest SemVer)">
    </a>
    <a href="https://github.com/rilldata/rill/commits" target="_blank">
        <img src="https://img.shields.io/github/commit-activity/y/rilldata/rill.svg" alt="GitHub commit activity">
    </a>
    <a href="https://github.com/rilldata/rill/graphs/contributors" target="_blank">
        <img src="https://img.shields.io/github/contributors-anon/rilldata/rill.svg" alt="GitHub contributors">
    </a>
    <a href="https://github.com/rilldata/rill/releases" target="_blank">
        <img src="https://img.shields.io/github/downloads/rilldata/rill/total.svg" alt="GitHub downloads">
    </a>
    <a href="https://github.com/rilldata/rill/actions/workflows/rill-cloud.yml" target="_blank">
        <img src="https://github.com/rilldata/rill/actions/workflows/rill-cloud.yml/badge.svg" alt="CI/CD">
    </a>
</p>

<div align="center">

[Docs](https://docs.rilldata.com/) • [Install](https://docs.rilldata.com/home/install) • [Quickstart](https://docs.rilldata.com/get-started/quickstart) • [Guides](https://docs.rilldata.com/guides) • [Reference](https://docs.rilldata.com/reference/project-files)

</div>

---

Rill delivers the fastest path from data lake to dashboard. **For data engineers and analysts**, it provides rapid, self-service dashboards built directly on raw data lakes, eliminating traditional BI complexity. **For data consumers**, it ensures reliable, fast-loading dashboards with accurate, real-time metrics.

Download Rill to start modeling data and create fast, exploratory dashboards in minutes:

```bash
curl https://rill.sh | sh
rill start my-rill-project
```

Unlike most BI tools, Rill comes with its own embedded in-memory database powered by DuckDB or ClickHouse. Data and compute are co-located, and queries return in milliseconds, so you can pivot, slice, and drill-down into your data instantly.

We also support bringing your own OLAP engine as a live connector with pushdown compute capabilities.

<p align="center">
  <img src="/docs/static/img/explore/dashboard101/multi-measure-select.png" alt="Rill dashboard example" width="80%">
</p>

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Rill's design principles:](#rills-design-principles)
- [Core Concepts](#core-concepts)
  - [BI-As-Code](#bi-as-code)
  - [Metrics Layer](#metrics-layer)
  - [AI Agents](#ai-agents)
- [Learn More](#learn-more)
- [Production Examples](#production-examples)
  - [Programmatic Ads/OpenRTB](#programmatic-adsopenrtb)
  - [Cost Monitoring](#cost-monitoring)
  - [GitHub Analytics](#github-analytics)
  - [App Engagement](#app-engagement)
  - [Kitchen-sink example](#kitchen-sink-example)
- [Get in touch!](#get-in-touch)
- [Company](#company)
- [License](#license)

## Rill's design principles:

- **Lightning Fast** - Powered by SvelteKit & DuckDB for conversationally fast performance
- **Universal Data Support** - Works with local and remote datasets (Parquet, CSV, S3, GCS, HTTPS, local)
- **Automatic Profiling** - Build intuition about your dataset through automatic profiling
- **Real-time Response** - Responds to each keystroke by re-profiling the resulting dataset
- **Interactive Dashboards** - Thoughtful, opinionated defaults for quick insights
- **Dashboards as Code** - Version control, Git sharing, and easy project rehydration

## Core Concepts

### BI-As-Code

Rill implements BI-as-code through a combination of:

1. **SQL-based Definitions**: Define your models via SQL to connect to your various sources
2. **YAML Configuration**: Configure your metrics views, dashboards, and project settings via YAML
3. **Git Integration**: Version control your analytics assets
4. **CLI Tools**: Deploy and manage your analytics stack from the command line

<p align="center">
  <img src="https://docs.rilldata.com/img/concepts/metrics-view/metrics-view-components.png" alt="Rill Fundamentals" width="80%">
</p>

### Metrics Layer

Rill's metrics layer provides a unified way to define, compute, and serve business metrics. Metrics views combine SQL models with YAML configuration to create standardized, reusable business metrics that can be consumed by dashboards, APIs, and AI systems.

Example Metrics View:

```yaml
# metrics/revenue_metrics.yaml
name: revenue_metrics
description: Key revenue metrics by country and product
model: revenue_model
timeseries: date
dimensions:
  - name: country
    column: country
  - name: product_category
    column:product_category
measures:
  - name: total_revenue
    expression: sum(amount)
    description: Total revenue amount
  - name: order_count
    expression: count(*)
    description: Number of orders
  - name: avg_order_value
    expression: sum(amount) / count(*)
    description: Average order value
```

### AI Agents

We understand the critical importance of AI and data in modern business intelligence. Our metrics layer is designed to provide AI systems with the structured, real-time data they need to deliver quick and accurate responses. By co-locating data and compute with embedded databases like DuckDB and ClickHouse, Rill eliminates the latency that traditional BI tools introduce, ensuring AI agents can access fresh metrics instantly for precise decision-making and intelligent automation.

## Learn More

For visual learners, take a look at our various playlists that explains what Rill is and how to get the most out of it!

<div align="center">

[Getting Started with Rill Developer](https://www.youtube.com/watch?v=oQSok8Dy-D0) • [Exploring Data with Rill](https://www.youtube.com/watch?v=wTP46eOzoCk&list=PL_ZoDsg2yFKgi7ud_fOOD33AH8ONWQS7I&index=1)
• [Data Talks on the Rocks](https://www.youtube.com/playlist?list=PL_ZoDsg2yFKgr_YEc4XOY0wlRLqzyR07q)

</div>

## Production Examples

### Programmatic Ads/OpenRTB

Bidstream data for programmatic advertisers to optimize pricing strategies, look for inventory opportunities, and improve campaign performance.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads">GitHub →</a><br />
- <a href="/guides/openrtb-analytics">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-openrtb-prog-ads">Live Demo →</a>

### Cost Monitoring

Based on Rill's own internal dashboards, cloud infrastructure data (compute, storage, pipeline statistics, etc.) merged with customer data to analyze bottlenecks and look for efficiencies.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring">GitHub →</a><br />
- <a href="/guides/cost-monitoring-analytics">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-cost-monitoring">Live Demo →</a>

### GitHub Analytics

Analyze GitHub activity to understand what parts of your codebase are most active, analyze contributor productivity, and evaluate the intersections between commits and files.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics">GitHub →</a><br />
- <a href="/guides/github-analytics">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-github-analytics">Live Demo →</a>

### App Engagement

A conversion dataset used by marketers, mobile developers, or product teams to analyze funnel steps.

- <a href="https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement">GitHub →</a><br />
- <a href="https://ui.rilldata.com/demo/rill-app-engagement">Live Demo →</a>

### Kitchen-sink example

A compilation of projects with deep dives into Rill's features using ClickHouse's GitHub commit information.

- <a href="https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial">GitHub →</a><br />
- <a href="/guides/rill-basics/launch">Walkthrough →</a><br />
- <a href="https://ui.rilldata.com/demo/my-rill-tutorial">Live Demo →</a>

## Get in touch!

- **[Discord Community](https://discord.gg/2ubRfjC7Rh)** - Join discussions and get help
- **[GitHub Issues](https://github.com/rilldata/rill/issues)** - Report bugs and request features
- **[Rill Guru](https://gurubase.io/g/rill)** - Ask questions and get expert answers

## Company

Rill is developed and maintained by [Rill Data, Inc.](https://www.rilldata.com/).

## License

This project is licensed under the [Apache License 2.0](LICENSE) - see the [LICENSE](LICENSE) file for details.
