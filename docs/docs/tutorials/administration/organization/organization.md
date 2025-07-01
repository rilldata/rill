---
title: "Organization Management"
sidebar_label: "Organization Management"
sidebar_position: 1
hide_table_of_contents: false
tags:
  - Administration
  - Organization
---

# Organization Management in Rill

Rill provides comprehensive organization management capabilities that allow you to structure your analytics environment, manage users, and customize the icons and logo of your environment. This guide covers the key concepts and features for managing your Rill organization.

## What is an Organization?

An organization in Rill is the top-level container that groups related projects, users, and resources. It serves as the foundation for:

- **User Management**: Adding, removing, and managing user access
- **Project Organization**: Grouping related projects and dashboards
- **Billing**: Managing subscription and usage costs, [Team plans](/manage/account-management/billing#team-plan)

## Organization Structure

### Organization Hierarchy

```
Organization
├── Users & Groups
│   ├── Admins
│   ├── Editors
│   ├── Members
│   └── Guests
├── Projects
│   ├── Project A
│   └── Project B
└── Resources
    ├── Data Sources
    ├── Models
    ├── Metrics Views
    ├── Dashboards
    └── APIs
```

### Key Components

#### 1. Organization Settings
- **Customization**: Organization name, description, Logo and branding
- **Domain**: Custom domain configuration, [contact us](/contact) for more information!
- **Billing**: Subscription management and usage tracking

#### 2. [Project Management](/tutorials/administration/project/project-maintanence)
- **Environment Variable Management**: Development, staging, and production environments
- **Resource Status Moniitoring**: Managing model refresh, initiate full project refresh, GitHub Repository
- **Setup MCP Connectivity**: Connect to your Agent of choice via the AI tab. 
- **Rill Resource Management**: Manage your Reports, Alerts, and Public URLs.

#### 3. [User Management](/tutorials/administration/user/user-management)
- **User Roles**: Admin, Member, Editor, Guest with different permission levels
- **User Groups**: Logical groupings of users for easier access management

## Getting Started with Organization Management

### 1. Creating an Organization

When you first deploy to Rill, we will extrapolate the organization name (usually based on the domain or email used to register), unless defined explicitly. This is either done from the UI deployment (if you followed the tutorial) or directly from the [CLI](/deploy/deploy-dashboard/#deploying-a-project-via-the-cli).



### 2. Managing an Organization

Once deployed, you can manage the organization via the Users or Settings page. Alternatively, the CLI also allows you to make similar modifications.

```bash
rill org

Manage organisations

Usage:
  rill org [command]

Available Commands:
  create         Create organization
  edit           Edit organization details
  switch         Switch to other organization
  list           List all organizations
  show           Show org details
  delete         Delete organization
  rename         Rename organization
  upload-logo    Upload a custom logo
  upload-favicon Upload a custom favicon

Global Flags:
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)

Use "rill org [command] --help" for more information about a command.
```
