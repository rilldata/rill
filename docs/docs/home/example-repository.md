---
title: Example Repository
sidebar_label: Example Repository
sidebar_position: 15
---

Explore our [public example repository](https://github.com/rilldata/rill-examples/) to jumpstart your Rill journey with real-world projects and use cases.

<img src = '/img/tutorials/rill_basics/new-rill-project.png' class='rounded-gif' />
<br />

Every example project includes comprehensive documentation covering:

- **Data Sources & Models**: Detailed information about source data and any required data modeling
- **Metrics & Dimensions**: Complete definitions of metrics and dimensions used in metrics views
- **Dashboards & Analysis**: Working examples of Explore and Canvas dashboards for data analysis

## Featured Example Projects

Our repository contains several production-ready examples:

- **[App Engagement](https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement)**: Conversion funnel analysis for marketers, mobile developers, and product teams to track user journey optimization
- **[Cost Monitoring](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring)**: Cloud infrastructure analytics combining compute, storage, and pipeline metrics with customer data to identify bottlenecks and optimize resource efficiency
- **[GitHub Analytics](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics)**: Repository activity analysis to identify codebase hotspots, measure contributor productivity, and track commit-file relationships. [View walkthrough →](/guides/github-analytics)
- **[Programmatic Ads/OpenRTB](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads)**: Bidstream analytics for programmatic advertisers to optimize pricing strategies, discover inventory opportunities, and enhance campaign performance. [View walkthrough →](/guides/openrtb-analytics)
- **[Guided Tutorial Project](https://github.com/rilldata/rill-examples/tree/main/my-rill-tutorial)**: A comprehensive tutorial project showcasing Rill's latest features with practical examples, regularly updated with new capabilities. [View walkthrough →](/guides/tutorial/rill-basics/launch)

## Getting Started

### Prerequisites

Install Rill using our one-line installer:

```bash
curl https://rill.sh | sh
```

### Running an Example

Clone the repository and launch any example project:

```bash
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads
rill start
```

Rill will automatically build your project from data sources to dashboards and open it in your browser.

