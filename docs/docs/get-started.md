---
title: Get started
sidebar_label: Get started
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

This tutorial is a step by step guide that will help you install Rill, ingest source data, and deploy a dashboard online.

## Install Rill

Install `rill` on Mac or Linux ([WSL](https://learn.microsoft.com/en-us/windows/wsl/install) for Windows users) using the installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

## Create your project

Start a new Rill project:

```
rill start my-rill-project
```

The Rill web app runs locally at `http://localhost:9009` and will create code files in the `my-rill-project` directory.

## Load and transform data

On the welcome screen, initialize an example project or load up Rill with your own data. Use the app to develop data models and dashboards.

## Deploy your dashboard

You can deploy any Rill project with a dashboard to an authenticated hosted endpoint:

1. Create a new Github repository on [https://github.com/new](https://github.com/new) and push your `my-rill-project` directory to it
2. Setup continuous deployment from Github to Rill Cloud:
    ```
    cd my-rill-project
    rill deploy
    ```

## Share your dashboard

Projects on Rill Cloud are private by default. To invite others to explore your project, run:
```
rill user add
```

## We want to hear from you

You can file an issue [on GitHub](https://github.com/rilldata/rill/issues/new/choose) or reach us in our [Discord channel](https://bit.ly/3unvA05).
