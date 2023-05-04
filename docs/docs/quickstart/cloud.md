---
title: Quick start on cloud
sidebar_label: Cloud
sidebar_position: 20
---

Follow this tutorial to clone a starter project on Github and deploy it to Rill Cloud in five minutes.

## Introduction

In Rill, all data sources, data models, and dashboard definitions are defined using YAML and SQL code files. You can edit these files using Rill or directly using a code editor of you choice.

Rill Cloud connects to a repository on Github containing a Rill project, and continuously deploys that project on every push.

## Clone the starter project

For this tutorial, you will clone and deploy the Rill starter project on Github:

1. Head over to [https://github.com/rilldata/rill-developer-example](https://github.com/rilldata/rill-developer-example)

2. Click "Use this template" and select "Create a new repository"

3. Enter a name and click "Create repository from template" to clone the template into your Github account

4. Use Git to clone your copy of the project to your local computer:
```
git clone https://github.com/<your-account>/<your-rill-example-clone>.git
cd <your-rill-example-clone>
```

## Install Rill

Next, install Rill to edit and deploy the project.

On macOS, we recommend installing `rill` using Homebrew:

```
brew install rilldata/tap/rill
```

On Linux, we recommend installing `rill` using our installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

See [Install options](../develop/install.md) for more information about different ways to install Rill.

## Deploy to Rill Cloud

Before editing the project, let's wire up your deployment. In the directory containing the cloned starter project, run:

```
rill deploy
```

The CLI will guide you through authenticating with Rill Cloud and granting read-only access to your Rill project on Github.

## Edit the project

Congratulations! You have now deployed a project to Rill Cloud. Every time you `git push` changes to the project, Rill Cloud will automatically re-deploy your project.

To edit your project, run the local Rill application:

```
rill start
```

When you run `rill start`, it parses your project and ingests any missing data sources into a local DuckDB database. After your project has been re-hydrated, it starts the Rill web app on `http://localhost:9009`.

You can now use the Rill web app to add or edit data sources, data models, and dashboards. Have fun building your Rill project!

When you're done, don't forget to push your changes:

```
git add .
git commit -m "Updated project"
git push
```

## Share the project

Projects on Rill Cloud are private by default. To invite others to explore your project on Rill Cloud, run:
```
rill user add
```

Alternatively, make your Rill project publicly accessible:
```
rill project edit --public=true
```

## We want to hear from you

You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Discord channel](https://bit.ly/3unvA05). 
