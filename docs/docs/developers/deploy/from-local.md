---
title: Local Deploy
sidebar_label: Local Deploy
sidebar_position: 10
---

import ThemedImage from '@theme/ThemedImage';

Deploy dashboards you've built locally with Rill Developer — whether you authored them yourself or with an AI coding agent. This uploads your project files to Rill Cloud, where your team can access dashboards, alerts, APIs, and more.

Two options:

- **Direct upload** — Push a snapshot of your project to Rill Cloud. You push updates manually. Good for solo work or getting started fast.
- **GitHub-connected** — Link a GitHub repo for continuous deployment. Best for teams and production workflows.

## Deploy with an AI agent

If you're using an AI coding agent like [Claude Code](https://claude.ai/code) or [Cursor](https://www.cursor.com/), you can ask it to deploy your project directly. The agent will run the same CLI commands described below on your behalf, read any errors from `rill project status`, and iterate — all without leaving your editor.

**Example prompts:**

- *"Deploy my Rill project to Rill Cloud."* → runs `rill project deploy`
- *"Connect my project to GitHub and set up continuous deployment."* → runs `rill project connect-github`
- *"Create a new GitHub repo for this project and connect it to Rill Cloud."* → runs `rill project connect-github` and handles repo creation
- *"Check my deployment status and fix any errors."* → runs `rill project status`

## Deploy without GitHub

The fastest way to get started. This uploads a snapshot of your local project files to Rill Cloud. Your project won't update automatically — run the deploy command again or click **Update** whenever you want to push new changes. You can always [connect GitHub later](#connect-github-to-an-existing-project).

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

Already deployed without GitHub? You can add it anytime from the UI or CLI. Follow the steps in [Deploy with GitHub](#deploy-with-github) above — the process is the same whether you're deploying for the first time or connecting an existing project.

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
