---
title: Quick start
sidebar_label: Quick start
sidebar_position: 15
---

Follow this tutorial to install Rill and create a project on your local computer in less than two minutes.

## Installation

On macOS, we recommend installing `rill` using Brew:

```bash
brew install rilldata/rill-developer/rill
```

On Linux, we recommend installing `rill` using our installation script:

```bash
curl -s https://cdn.rilldata.com/install.sh | bash
```

See [Install options](./sql-models.md) for more information about different ways to install Rill.

<!-- TODO: Add docs link here -->

## Creating a project

In Rill, all data sources, data models, and dashboard definitions are saved as code artifacts on disk. You can edit these directly or check them into Git to share your project with others.

For this tutorial, let's checkout an example project from the git repository:

```bash
git clone https://github.com/rilldata/rill-developer.git
cd rill-developer/examples/sf_props
```

Alternatively, you can create a new, empty Rill project:

```bash
rill init --project my-project
cd my-project
```

## Starting the application

Now it's time to start the application:

```bash
rill start
```

When you run `rill start`, it parses your project and ingest any missing data sources into a local DuckDB database. After your project has been re-hydrated, it starts the Rill web app on `http://localhost:9009`.

You can now use the Rill web app to add or edit data sources, data models, and dashboards. All changes you make in the UI are reflected in the code artifacts stored on disk. You can share your Rill code artifacts with others, and they'll be able to completely recreate your project just by running `rill start`.

Have fun exploring Rill!

## We want to hear from you

You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Discord channel](https://bit.ly/3unvA05). Please abide by the [Rill Community Policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).
