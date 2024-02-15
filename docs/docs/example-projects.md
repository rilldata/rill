---
title: Example Projects
sidebar_label: Example Projects  
sidebar_position: 50
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Project Repository

We have created several example projects that highlight common use cases for Rill. 

The monorepo for these examples can be found at: https://github.com/rilldata/rill-examples/

Each example project includes a ReadMe with more details on:

- Source data in the dataset
- Dimension and metric definitions
- Example dashboard analyses

Current projects include:

- [App Engagement](https://github.com/rilldata/rill-examples/tree/main/rill-app-engagement): a conversion dataset used by marketers, mobile developers or product teams to analyze funnel steps
- [Cost Monitoring](https://github.com/rilldata/rill-examples/tree/main/rill-cost-monitoring): based off of Rill's own internal dashboards, cloud infrastructure data (compute, storage, pipeline statistics, etc.) merged with customer data to analyze bottlenecks and look for efficiencies
- [GitHub Analytics](https://github.com/rilldata/rill-examples/tree/main/rill-github-analytics): analyze GitHub activity to understand what parts of your codebase are most active, analyze contributor productivity, and evaluate the intersections between commits and files
- [Programmatic Ads/OpenRTB](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads): bidstream data for programmtic advertisers to optimize pricing strategies, look for inventory opportunities, and improve campaign performance
- [311 Operations](https://github.com/rilldata/rill-examples/tree/main/rill-311-ops): a live datastream of 311 call centers from various locations in the US for example operational analytics 


## Installing Examples

You can install `rill` using our installation script:

```
curl https://rill.sh | sh
```

To run an example (in this case our Programmatic/OpenRTB dataset):
```
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads
rill start
```

Rill will build your project from data sources to dashboard and then launch in a new browser window.
