---
title: Deploy Dashboards
sidebar_label: Deploy Dashboards
sidebar_position: 00
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

Once you've built your dashboards locally, deploying to Rill Cloud lets you share them with your team, set up [alerts and scheduled reports](/guide/dashboards/explore), [embed dashboards](/developers/integrate/embedding) in other apps, and collaborate with others.

- [Deploy from the UI](#deploy-from-the-ui) — The quickest way to get started
- [Deploy from the CLI](#deploy-from-the-cli) — For scripting and automation
- [Deploy with GitHub](#deploy-with-github) — Best for teams and ongoing projects
- [Connect GitHub later](#connect-github-to-an-existing-project) — Add version control after your first deploy
- [Managing access](#managing-access) — Invite your team

## Deploy from the UI

The fastest way to deploy is right from Rill Developer. Just click the **Deploy** button and follow the prompts.

<img src='/img/deploy/existing-project/deploy-ui.gif' class='rounded-gif' />
<br/>

Made some changes? Click **Update** to push them to the cloud.

<img src='/img/deploy/existing-project/redeploy.gif' class='rounded-gif' />

## Deploy from the CLI

Prefer the command line? Run:

```bash
rill project deploy
```

This uploads your project directly to Rill Cloud. It's great for quick deploys or when you want to set up GitHub integration later.

## Deploy from CLI with GitHub

For team projects or dashboards you'll maintain over time, we recommend connecting to GitHub. This gives you:

- **Automatic updates** — Push to GitHub, and your dashboards update automatically
- **Version history** — See what changed and roll back if needed
- **Team collaboration** — Everyone can contribute through pull requests
- **BI-as-code** — Your dashboards live alongside your other code

To deploy with GitHub from the start, run:

```bash
rill project connect-github
```

Rill will either create a new repository or connect to an existing one, then set up continuous deployment for you.

**Helpful options:**
- `--prod-branch [BRANCH]` — Deploy from a branch other than main
- `--subpath path/to/project` — If your Rill project lives inside a larger repo

:::tip New to Git?
No problem! Check out our [GitHub Basics](/developers/guides/deploy/github-101) guide, which walks you through everything using GitHub Desktop — no command line required.
:::

## Connect GitHub to an Existing Project

Already deployed without GitHub? You can add it anytime:

1. Go to the **Status** page in Rill Cloud
2. Click **Connect to GitHub**
3. Create a new repo or pick an existing one

<img src='/img/deploy/existing-project/install-rill-cloud.png' class='rounded-gif' />

:::note Need admin help?
The Rill Cloud GitHub app needs permission to read and write to your repository. If you're not a GitHub org admin, you may need to ask them to approve the app first.
:::

## Managing Access

Your dashboards are private by default. Once deployed, head to [User Management](/guide/administration/users-and-access/user-management) to invite your team and set up permissions.

## Something not working?

Check your deployment status anytime:

```bash
rill project status
```

You can also view detailed status information on the Status page in Rill Cloud.

:::tip Using GitLab?
We've got you covered. See [Deploy from CLI](/developers/guides/deploy-from-cli) for GitLab instructions.
:::
