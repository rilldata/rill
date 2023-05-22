---
title: Data to hosted dashboard
sidebar_label: Data to hosted dashboard
sidebar_position: 1
---

This tutorial is a step by step guide that will help you install Rill, ingest source data, and deploy a secure dashboard application online.


## Installation

You can install `rill` on Mac or Linux ([WSL](https://learn.microsoft.com/en-us/windows/wsl/install) for Windows users) using our installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

See [install options](../develop/install.md) for more information about different ways to install Rill.

## Creating a project

All data sources, data models, and dashboard definitions are saved as SQL and YAML files on disk. You can edit these directly or check them into git to share your project with others.

For this tutorial, let's checkout example project files from a Github repository:

```
git clone https://github.com/rilldata/rill-developer-example.git
cd rill-developer-example
```

Alternatively, you can create a new, empty Rill project and [add data that is meaningful to you](../develop/import-data) using the CLI or application:

```
rill init my-project
cd my-project
```


## Starting the application

Once inside the project directory, start the application:

```
rill start
```

When you run `rill start`, it parses your project and ingests any missing data sources into a local duckDB database. The Rill web app runs locally at `http://localhost:9009`.

## Editing and collaborating

The local Rill application helps you go from data to dashboard using SQL transformations and simple YAML configuration files.


### Project files and git
All changes you make in the application UI and [CLI](../reference/cli/cli.md) are versionable because they are reflected as [Rill project files](../reference/project-files) stored on disk. You can collaborate on your Rill project files with others by pushing to a shared repository. After cloning, your collaborator will be able to completely recreate your project locally by cloning the files running `rill start`.

### Data sources

Rill supports [several connectors for importing data](../develop/import-data): local files, download from an S3 or GCS bucket, or download using HTTP(S). Rill can ingest .csv, .tsv, and .parquet files, which may be compressed (.gz). 


### Transformations

Data [transformations in Rill](../develop/sql-models) are composed of SQL SELECT statements that model source data. They allow you to join, transform, and clean data. Data transformations in Rill Developer are powered by [duckDB](https://duckdb.org/docs/) and their dialect of SQL. Please visit their documentation for insight into how to write your queries.


### Dashboards
To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable metrics definition to enable ad hoc data exporation through Rill's interactive dashboard.

[Metrics are created](../develop/metrics-dashboard) on top of this table to create powerful interactive dashboards.

## Rill hosted dashboards

Rill hosted dashboards allow you to create an authenticated web application that can be shared with others. The hosted endpoint uses the output of your local project to create version of the project for _dashboard viewers_ to drill into. There is no SQL editing or dashboard configuration available in the hosted experience. This helps everyone explore data insights interactively without needing to touch code and configure settings.

## Deploy dashboards

You can [deploy](../deploy/existing-project) any Rill project with a valid dashboard to an authenticated hosted endpoint by running:

```
rill deploy
```

The CLI will guide you through authenticating with Rill and granting read-only access to your Rill project on Github.

### Github connection

Your project [_must_ be on Github](../deploy/existing-project) before you deploy it to Rill. Once Rill is connected to a valid project hosted on Github, we will continuously deploy `main` on every push to update your viewer's experience.


### Data connection

When running Rill locally, we attempt to find existing credentials configured on your computer. When deploying projects, you [must explicitly provide service account credentials](/deploy/configure-connector-credentials) with correct access permissions. Talk to your IT team if you do not know how to create a service account.


## Restarting your project locally
When you run `rill start`, it parses your project and ingests any missing data sources into a local duckDB database. After your project has imported all sources, it starts the Rill web app on `http://localhost:9009`.

You can now use the Rill web app to add or edit data sources, data models, and dashboards. 

When you're done, don't forget to push your changes and merge them into the main branch in Github to bring them to your hosted dashboard viewers:

```
git add .
git commit -m "Updated project"
git push
```

## Share hosted dashboard

Projects on Rill are private by default. To invite others to explore your project on Rill, run:
```
rill user add
```

Alternatively, make your Rill project publicly accessible:
```
rill project edit --public=true
```

## We want to hear from you

You can file an issue in [this repository](https://github.com/rilldata/rill-developer/issues/new/choose) or reach us in our [Discord channel](https://bit.ly/3unvA05). 
