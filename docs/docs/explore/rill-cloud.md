---
title: "What is Rill Cloud?"
description: Deploy, share, and collaborate on your Rill dashboards
sidebar_label: "Rill Cloud"
sidebar_position: 00
---

## Overview

Rill Cloud is a fully-managed platform for deploying, sharing, and collaborating on your Rill dashboards. While [Rill Developer](/get-started/install) is perfect for local development and exploration, Rill Cloud enables you to share your work with your team and stakeholders, manage access controls, and set up automated reports and alerts. For a more detailed document on the differences see [Rill Cloud vs Rill Developer](/get-started/concepts/cloud-vs-developer).

After logging into [Rill Cloud](https://ui.rilldata.com), you'll see all projects within your [organization](/manage/organization-management#organization) that are available and have been granted permissions to your user profile. Within each project, you can access the corresponding dashboards, manage alerts and reports, chat with your data using AI, and configure project settings.

<img src = '/img/manage/project-management/rill-org.png' class='rounded-gif' />
<br />

## Organization Structure

Rill Cloud is organized hierarchically to help you manage access and resources efficiently:

- **[Organizations](/manage/organization-management)** - The top-level container for your team. Organizations contain projects, users, groups, and billing settings.
- **[Projects](/manage/project-management)** - Individual Rill projects within an organization. Each project has its own data sources, models, metrics views, and dashboards.
- **[Users & Groups](/manage/user-management)** - Team members and their access permissions can be managed at the organization or project level using [roles and permissions](/manage/roles-permissions).
- **Settings** - Configure organization-wide settings like billing, branding (logo and favicon), and general preferences.

:::tip Quick Start
New to Rill Cloud? Check out our [deployment guide](/deploy/deploy-dashboard) to learn how to deploy your first project, or try one of our [demo projects](https://ui.rilldata.com/demo) to see what Rill Cloud can do.
:::

## Rill Cloud Project Features

Each project in Rill Cloud comes with a comprehensive set of features for exploring, monitoring, and sharing your data:

### AI-Powered Chat
Ask questions about your data in natural language using AI. Rill offers two ways to interact with your data conversationally:

- **[AI Chat (Project Chat)](/explore/project-chat)** - Built directly into Rill Cloud, ask questions about your metrics and get instant insights without writing queries.
- **[Rill MCP Server](/explore/mcp)** - Connect your Rill projects to AI assistants like Claude Desktop using the Model Context Protocol for governed, accurate analytics.

### Dashboards
Explore your data through interactive dashboards that make it easy to slice, dice, and drill down into your metrics:

- **[Dashboard Quickstart](/explore/dashboard-101)** - Learn the basics of navigating and using Rill dashboards with measures, dimensions, and time series.
- **[Filters & Comparisons](/explore/filters)** - Apply powerful filters and time comparisons to focus your analysis.
- **[Bookmarks](/explore/bookmarks)** - Save specific dashboard states (filters, metrics, dimensions) and share them with others.
- **[Public URLs](/explore/public-url)** - Share dashboards externally with customers and partners without requiring them to have Rill accounts.

### Reports
Set up automated data exports and scheduled email reports to keep your team informed:

- **[Exports & Scheduled Reports](/explore/exports)** - Export data in CSV, Excel, or Parquet formats, or schedule recurring reports to be delivered to your inbox.

Reports are managed from your project home page under the **Reports** tab. You can view, edit, and delete scheduled reports, and see execution history for all deliveries.

### Alerts
Stay on top of important changes in your data with automated alerting:

- **[Alerts](/explore/alerts/alerts.md)** - Create alerts on any measure with custom criteria and thresholds. Get notified via email or Slack when conditions are met.
- **[Slack Integration](/explore/alerts/slack)** - Connect your Slack workspace to receive alert notifications in channels or direct messages.

Alerts are accessible from any dashboard via the bell icon in the upper-right corner, and can be managed from the **Alerts** tab on your project home page.

### Status
Monitor the health and performance of your project in real-time:

- View data source refresh status and history
- Check model build status and execution logs  
- Monitor query performance and resource usage
- Review project deployment history

The Status page helps you quickly identify and troubleshoot any issues with your data pipeline or dashboards.

### Settings
Configure your project with credentials, variables, and access controls:

- **[Variables & Credentials](/manage/project-management/variables-and-credentials)** - Store sensitive credentials and configure environment-specific variables securely.
- **[User Management](/manage/user-management)** - Invite team members and configure their access at the project level.
- **Project Details** - Edit project name, description, and other metadata.