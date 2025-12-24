---
title: "Getting Started with Rill Cloud"
description: "Introduction to Rill Cloud, AI features, and management"
sidebar_label: "Get Started"
sidebar_position: 0
slug: "/user-guide/getting-started"
---

# Getting Started with Rill Cloud

Welcome to Rill Cloud! This guide will help you understand the core concepts and features available in Rill Cloud, from exploring your data with AI to managing your organization and projects.

## Why Rill Cloud?

Business users and data analysts need fast, interactive access to their data—without waiting for IT teams to build custom reports or learning complex SQL. Rill Cloud brings powerful analytics directly to your fingertips, enabling you to explore, analyze, and share insights at the speed of thought.

### What Makes Rill Cloud Different?

#### Fast, Interactive Exploration

Rill Cloud dashboards respond instantly, even with millions of rows of data. Slice, dice, and drill down into your metrics without waiting for queries to process. Change filters, compare time periods, and explore dimensions—all in real-time.

#### No SQL Required

Ask questions about your data in plain English using [AI Chat](/user-guide/ai/ai-chat). Get instant answers without writing queries or waiting for someone else to build a report. Rill understands your business metrics and provides context-aware insights.

#### Self-Service Analytics

Explore your data independently without depending on data teams. Dashboards are pre-configured with the metrics and dimensions that matter to your business, so you can focus on finding insights rather than building queries.

#### Easy Sharing and Collaboration

Share specific dashboard views with [bookmarks](/user-guide/dashboards/bookmarks), send [scheduled reports](/user-guide/reports/exports) to your inbox, and set up [alerts](/user-guide/alerts) to stay informed when metrics change. Collaborate with your team on insights without complex permission setups.

### How Rill Cloud Works for You

Rill Cloud provides a unified platform where your data team has already set up the metrics and dashboards you need. You simply:

1. **Explore dashboards** - Navigate interactive dashboards with measures, dimensions, and time series visualizations
2. **Ask questions** - Use AI Chat to get answers about your data in natural language
3. **Share insights** - Create bookmarks and share them with stakeholders
4. **Stay informed** - Set up alerts and scheduled reports for the metrics you care about

### Built for Business Users

Rill Cloud is designed for people who need to make data-driven decisions quickly:

- **Product Managers** - Track feature adoption, user engagement, and product metrics in real-time
- **Marketing Teams** - Analyze campaign performance, conversion funnels, and attribution without waiting for reports
- **Operations Teams** - Monitor system health, performance metrics, and operational KPIs as they happen
- **Business Analysts** - Explore data relationships, identify trends, and answer ad-hoc questions independently

Unlike traditional BI tools that require pre-built reports or SQL knowledge, Rill Cloud gives you the flexibility to explore your data interactively and discover insights on your own timeline.

### Key Benefits

- **Speed**: Get answers instantly with sub-second query performance, even on large datasets
- **Simplicity**: No SQL or technical skills required—use natural language to explore your data
- **Flexibility**: Explore data interactively without being limited to pre-built reports
- **Collaboration**: Share insights easily with bookmarks, reports, and alerts
- **AI-Powered**: Leverage AI to understand your data and get intelligent recommendations
- **Always Up-to-Date**: Dashboards automatically refresh with the latest data from your sources

## What is Rill Cloud?

Rill Cloud is a fully-managed platform for deploying, sharing, and collaborating on your Rill dashboards. While [Rill Developer](/developer/get-started/install) is perfect for local development, Rill Cloud enables you to:

- **Share dashboards** with your team and stakeholders
- **Collaborate** on data insights across your organization
- **Manage access controls** with fine-grained permissions
- **Set up automated reports** and alerts
- **Use AI-powered analysis** to explore your data conversationally

For a detailed comparison, see [Rill Cloud vs Rill Developer](/developer/deploy/cloud-vs-developer). 

## Understanding Rill Cloud Structure

Rill Cloud is organized hierarchically to help you manage access and resources efficiently:

### Organizations

**[Organizations](/user-guide/administration/organization-settings)** are the top-level container for your team. They contain:
- Projects and their dashboards
- Users and user groups
- Billing and subscription settings
- Organization-wide branding and preferences

When you log into [Rill Cloud](https://ui.rilldata.com), you'll see all projects within your organization that you have access to.

### Projects

**[Projects](/user-guide/administration/project-settings)** are individual Rill deployments within an organization. Each project contains:
- Data sources and connectors
- Models and transformations
- Metrics views and dashboards
- Project-specific settings and credentials
- User access controls

### Users and Access

Access in Rill Cloud is managed through:
- **[Users](/user-guide/administration/users-and-access/user-management)** - Individual team members with specific roles
- **[User Groups](/user-guide/administration/users-and-access/usergroup-management)** - Collections of users for easier permission management
- **[Roles and Permissions](/user-guide/administration/users-and-access/roles-permissions)** - Define what users can do (view, edit, admin)

## AI-Powered Features

Rill Cloud includes powerful AI capabilities to help you explore and understand your data:

### AI Chat

**[AI Chat](/user-guide/ai/ai-chat)** is built directly into Rill Cloud, allowing you to:
- Ask questions about your data in natural language
- Get instant insights without writing SQL queries
- Explore metrics and dimensions conversationally
- Receive explanations and recommendations

AI Chat understands your project's metrics views, measures, and dimensions, providing context-aware answers based on your data structure.

### Rill MCP Server

The **[Rill MCP Server](/user-guide/ai/mcp)** connects your Rill projects to external AI assistants like Claude Desktop using the Model Context Protocol. This enables you to:
- Query your Rill data from Claude Desktop or other MCP-compatible tools
- Get governed, accurate analytics through AI assistants
- Maintain security and access controls while using external AI tools
- Leverage advanced AI capabilities for data analysis

## Exploring Your Data

Rill Cloud provides powerful tools for exploring and analyzing your data:

### Interactive Dashboards

**[Dashboards](/user-guide/dashboards/dashboard-101)** are the primary way to explore your data in Rill Cloud:
- **Slice and dice** your metrics by dimensions
- **Filter** data to focus on specific subsets
- **Compare** time periods to identify trends
- **Drill down** into specific dimensions or time periods
- **Create pivot tables** for cross-tabulation analysis

### Bookmarks and Sharing

- **[Bookmarks](/user-guide/dashboards/bookmarks)** - Save specific dashboard views (filters, metrics, dimensions) and share them with your team
- **[Public URLs](/user-guide/dashboards/public-url)** - Share dashboards externally with customers and partners without requiring Rill accounts

### Reports and Alerts

- **[Scheduled Reports](/user-guide/reports/exports)** - Set up automated email reports with your key metrics
- **[Alerts](/user-guide/alerts)** - Get notified when metrics meet certain conditions via email or Slack

## Managing Your Organization

Rill Cloud provides comprehensive management tools for administrators:

### User Management

- **[User Management](/user-guide/administration/users-and-access/user-management)** - Invite team members, assign roles, and manage access
- **[User Groups](/user-guide/administration/users-and-access/usergroup-management)** - Organize users into groups for easier permission management
- **[Roles and Permissions](/user-guide/administration/users-and-access/roles-permissions)** - Understand and configure what users can do

### Organization Settings

- **[Organization Management](/user-guide/administration/organization-settings)** - Configure organization-wide settings, billing, and branding
- **[Project Management](/user-guide/administration/project-settings)** - Manage project settings, credentials, and variables
- **[Service Tokens](/user-guide/administration/access-tokens/service-tokens)** - Create tokens for programmatic access

## Additional Resources

- **[Demo Projects](https://ui.rilldata.com/demo)** - Explore live examples of Rill Cloud dashboards
- **[YouTube Playlist](https://www.youtube.com/watch?v=wTP46eOzoCk&list=PL_ZoDsg2yFKgi7ud_fOOD33AH8ONWQS7I&index=1)** - Video tutorials to get started with Rill Cloud
