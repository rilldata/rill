# Infrastructure Cost Monitoring
An example project that explores how Rill can be used to monitor cloud infrastructure over time. 

If you have added the full Rill Example project, run `rill start` from this directory to get started.

To run this example specifically:

```
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-cost-monitoring
rill start
```

Rill will build your project from data sources to dashboard and then launch in a new browser window.

## Overview
This dataset is modeled after a similar dashboard we use internally at Rill to both identify opportunities to improve our cloud infrastructure operations and to manage customer implementations. Typical users would include engineering, customer success and finance. In this example, we’ve tied together a combination of cloud services, other hosting costs, and revenue metrics.

## Data Model
In this dataset, you’ll see:

Key Dimensions:
- SKU (cloud provider services)
- Pipeline (internal data pipeline name)
- App Name (cloud provider application)

Key Metrics: 
- Revenue and cost

## Dashboard Details 
Some typical analyses that we use the monitoring dashboard for:
- Compare margin across customers to understand if a particular implementation can be enhanced 
- Trend cloud services at a highly granular level to understand revenue leakage 
- Quickly drill in on spikes in usage to identify opportunities for optimization
