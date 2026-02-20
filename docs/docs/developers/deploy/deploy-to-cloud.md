---
title: Deploy to Rill Cloud
sidebar_label: Deploy to Rill Cloud
sidebar_position: 00
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->
import ThemedImage from '@theme/ThemedImage';

Once you've built your dashboards locally, deploying to Rill Cloud lets you share them with your team, set up [alerts and scheduled reports](/guide/dashboards/explore), [embed dashboards](/developers/integrate/embedding) in other apps, and collaborate with others.

## Deploy without GitHub

The fastest way to get started. You can always [connect GitHub later](#connect-github-to-an-existing-project).

### From the UI

Click the **Deploy** button in Rill Developer and follow the prompts.

<img src='/img/deploy/existing-project/deploy-ui.gif' class='rounded-gif' />
<br/>

Made some changes? Click **Update** to push them to the cloud.

<ThemedImage
  alt="Update button in Rill Developer showing how to push local changes to Rill Cloud"
  sources={{
    light: '/img/deploy/update-light.png',
    dark: '/img/deploy/update-dark.png',
  }}
/>

### From the CLI

```bash
rill project deploy
```

This uploads your project directly to Rill Cloud. It's great for quick deploys or when you want to script deployments.

## Deploy with GitHub

Best for teams — gives you version control, continuous deployment, and PR workflows.

- **Automatic updates** — Push to GitHub, and your dashboards update automatically
- **Version history** — See what changed and roll back if needed
- **Team collaboration** — Everyone can contribute through pull requests
- **BI-as-code** — Your dashboards live alongside your other code

### From the UI

1. Deploy your project using the **Deploy** button
2. Go to the **Status** page in Rill Cloud
3. Click **Connect to GitHub**
4. Create a new repo or pick an existing one

<img src='/img/deploy/existing-project/install-rill-cloud.png' class='rounded-gif' />

:::note Need admin help?
The Rill Cloud GitHub app needs permission to read and write to your repository. If you're not a GitHub org admin, you may need to ask them to approve the app first.
:::

### From the CLI

```bash
rill project connect-github
```

Rill will either create a new repository or connect to an existing one, then set up continuous deployment for you.

:::tip New to Git?
No problem! Check out our [GitHub Basics](/developers/tutorials/github-101) guide, which walks you through everything using GitHub Desktop — no command line required.
:::

#### Deploy from a specific branch

By default, Rill deploys from the repository's default Git branch. To deploy from a different branch:

```bash
rill project connect-github --primary-branch my-branch-name
```

#### Deploy from a monorepo

If your Rill project lives inside a larger repository, use the `--subpath` flag to point to the project directory:

```bash
rill project connect-github --subpath path/to/project
```

:::warning
You must run `rill project connect-github` from the root of your Git repository, **not** the root of your Rill project.
:::

## Connect GitHub to an existing project

Already deployed without GitHub? You can add it anytime:

1. Go to the **Status** page in Rill Cloud
2. Click **Connect to GitHub**
3. Create a new repo or pick an existing one

<img src='/img/deploy/existing-project/install-rill-cloud.png' class='rounded-gif' />

:::note Need admin help?
The Rill Cloud GitHub app needs permission to read and write to your repository. If you're not a GitHub org admin, you may need to ask them to approve the app first.
:::

## Something not working?

Check your deployment status anytime:

```bash
rill project status
```

You can also view detailed status information on the Status page in Rill Cloud.

:::tip Using GitLab?
We've got you covered. See [Deploy from CLI](/developers/tutorials/deploy-from-cli) for GitLab instructions.
:::

Once deployed, your project is private by default. Head to [User Management](/guide/administration/users-and-access/user-management) to invite your team and set up permissions.
