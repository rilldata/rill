---
title: Get started
sidebar_label: Get started
sidebar_position: 10
---

This tutorial is a step by step guide that will help you install Rill, ingest source data, and deploy a secure dashboard application online.


## Install Rill

You can install `rill` on Mac or Linux ([WSL](https://learn.microsoft.com/en-us/windows/wsl/install) for Windows users) using our installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

## Create your project

Checkout our example project files from Github and start the project with Rill:

```
git clone https://github.com/rilldata/rill-developer-example.git
rill start rill-developer-example
```
The Rill web app runs locally at `http://localhost:9009`.


## Transform your data

Data [transformations in Rill](../develop/sql-models) are composed of SQL SELECT statements that model source data.


## Deploy your dashboard

You can [deploy](../deploy/existing-project) any Rill project with a dashboard to an authenticated hosted endpoint by running:

```
rill deploy
```


## Share your dashboard

Projects on Rill are private by default. To invite others to explore your project, run:
```
rill user add
```
## We want to hear from you

You can file an issue in [this repository](https://github.com/rilldata/rill-developer/issues/new/choose) or reach us in our [Discord channel](https://bit.ly/3unvA05). 
