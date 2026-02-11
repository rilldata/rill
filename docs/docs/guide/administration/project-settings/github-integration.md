---
title: "GitHub Integration"
description: "Connect and manage GitHub repository integration for your Rill Cloud projects"
sidebar_label: "GitHub Integration"
sidebar_position: 21
---

# GitHub Integration

Each Rill Cloud project can be connected to a single GitHub repository. This enables continuous deployment—your project automatically redeploys whenever you push changes to the connected repository.

## Connecting to a GitHub Repository

On first deployment, if you've deployed via the UI, your project will not be connected to a GitHub repository. You will need to manually connect it:

1. Navigate to your project's **Status** page
2. Select **Connect to GitHub**
3. Follow the steps to connect your repository

![Select Repo](/img/deploy/existing-project/select-repo.png)

## Modifying GitHub Repository Connection

In some cases, you will need to change the repository that your project is synced to:

1. Navigate to your project's **Status** page
2. Select the dropdown next to the repository name
3. Choose **Disconnect** to remove the current connection

![Disconnect Github](/img/manage/project-management/disconnect-github.png)

This action has no effect on your current deployment and will not require a source re-ingest. After disconnecting, you can follow the same steps as [connecting to a GitHub repository](#connecting-to-a-github-repository) to re-connect your project to a new repository.

## Deploying from a Branch Other Than `main`

By default, Rill Cloud deploys from the `main` branch of your connected repository. You can change this to deploy from a different branch.

### Via Rill Cloud UI

If you have already [setup your connection to GitHub](/developers/deploy/deploy-dashboard/#connect-github-to-an-existing-project), you can edit the branch from the project settings:

![Main Branch](/img/manage/project-management/main-branch.png)

### Via CLI

To change the branch via the CLI, run:

```bash
rill project edit
```

This will open an interactive prompt where you can update the branch name and other project properties.

## Automatic Deployment

Your project on Rill Cloud will automatically redeploy every time you push changes to the connected GitHub repository. This ensures your dashboards always reflect the latest version of your project code.

:::tip Manual Refresh
To manually refresh data sources without pushing code changes (or redeploying your project), use:
```bash
rill project refresh [--source/model] (source_name or model_name)
```
:::

