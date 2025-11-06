---
title: What is BI-as-code?
sidebar_label: What is BI-as-code?
sidebar_position: 13
hide_table_of_contents: true
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Overview

BI-as-code is a modern approach to business intelligence that treats analytics assets as code, bringing the same benefits of version control, collaboration, and automation that software development teams have enjoyed for years. With Rill, you can define your entire analytics stack—from data models to dashboards—using code, while still maintaining the flexibility to make UI-based adjustments when needed.

## How It Works in Rill

Rill implements BI-as-code through a combination of:

1. **SQL-based Definitions**: Define your models via SQL to connect to your various sources
2. **YAML Configuration**: Configure your metrics views, dashboards, and project settings via YAML
3. **Git Integration**: Version control your analytics assets
4. **CLI Tools**: Deploy and manage your analytics stack from the command line
   
<div style={{ textAlign: 'center' }}>
  <img src="/img/concepts/metrics-view/metrics-view-components.png" style={{ width: '100%', borderRadius: '15px', padding: '20px' }} />
</div>


This approach allows engineering teams to maintain control over their analytics stack while enabling business users to make adjustments through the UI when needed.

## Key Benefits
 
### [Version Control & Collaboration](/deploy/deploy-dashboard)
- Track changes to your analytics assets in Git
- Review and approve changes through pull requests
- Maintain a clear history of how metrics and dashboards evolve over time
- Collaborate effectively across teams

### [Automation & CI/CD](/deploy/deploy-dashboard/github-101)
- Automate the deployment of analytics changes
- Integrate analytics testing into your CI/CD pipeline
- Ensure consistency across environments
- Reduce manual deployment errors

### Developer Experience
- Use familiar tools and workflows (Git, CLI, IDEs)
- Write SQL for last-mile ETL
- Maintain analytics as part of your codebase
- Leverage existing development practices

### Flexibility & Coexistence
- Define [core metrics and dimensions in code](/build/metrics-view)
- Make UI-based adjustments when needed
- Both code-defined and UI-created assets live in harmony
- Best of both worlds: developer control and business agility


## Learn More

For a deeper dive into BI-as-code and its benefits, check out our blog post: [What is BI-as-code?](https://www.rilldata.com/blog/bi-as-code-and-the-new-era-of-genbi)

## Next Steps

- [Learn about Rill's Architecture](/get-started/concepts/architecture)
- [Get started with Rill](/get-started/install)
- [Explore the Reference](/build/connectors)
- [Step-by-step Tutorial](/guides)
