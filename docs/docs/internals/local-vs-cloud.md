
---
title: Local vs. Cloud
description: Explaining Rill's local and cloud environments
sidebar_label: Local vs. Cloud
sidebar_position: 20
---

# Local vs. Cloud

Rill has two complementary environments:

## Rill Developer (Local)

The local CLI application (`rill`) for building projects.

The most notable command is `rill start <path>`, which starts a long-running process that:
- Watches files and re-executes on change
- Serves an IDE at `http://localhost:9009`
- Auto-syncs environment variables with Rill Cloud (if authenticated)

Additionally, the `rill` CLI can authenticate with Rill Cloud, storing an access token in `~/.rill`. This unlocks a variety of management commands for Rill Cloud, including:
- Managing orgs, projects, users, usergroups and service accounts
- Changing roles and access restrictions
- Viewing project status, logs, and resource state
- Triggering project refreshes, including granular refresh of specific resources or model partitions

## Rill Cloud

The managed cloud service for production deployments. Usually accessed on `ui.rilldata.com` or `api.rilldata.com`.

**Deployment options:**
- Connect a GitHub repository for continuous deploys on push
- Manual deploy from CLI or local IDE

**Production features:**
- User management and RBAC
- Data orchestration at scale
- Dashboard serving and monitoring
- UI for creating alerts and reports
- Link-based sharing and embedding

## Integration Points

The local and cloud environments connect at several points:
- **Authentication**: Local CLI authenticates with Cloud via OAuth
- **Environment variables**: `rill start` automatically syncs env vars with the connected Cloud project
- **Project identification**: Cloud project is identified by the Git remote in the local directory
- **Cloud operations**: Local CLI can manage Cloud projects and user access (`rill -h` for details)
