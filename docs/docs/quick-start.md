---
title: Quick start
sidebar_label: Quick start
sidebar_position: 15
---

Follow this tutorial to install Rill and create a project on your local computer in less than two minutes.

## Installation

On macOS, we recommend installing `rill` using Homebrew:

```
brew install rilldata/tap/rill
```

On Linux, we recommend installing `rill` using our installation script:

```
curl -s https://cdn.rilldata.com/install.sh | bash
```

See [Install options](./using-rill/install.md) for more information about different ways to install Rill.

<!-- TODO: Add docs link here -->

## Creating a project

In Rill, all data sources, data models, and dashboard definitions are saved as Rill project files on disk. You can edit these directly or check them into Git to share your project with others.

For this tutorial, let's checkout an example project from the git repository:

```
git clone https://github.com/rilldata/rill-developer-example.git
cd rill-developer-example
```

Alternatively, you can create a new, empty Rill project:

```
rill init --project my-project
cd my-project
```

## Starting the application

Now it's time to start the application:

```
rill start
```

When you run `rill start`, it parses your project and ingests any missing data sources into a local DuckDB database. After your project has been re-hydrated, it starts the Rill web app on `http://localhost:9009`.

## Editing and sharing a project

You can now use the Rill web app to add or edit data sources, data models, and dashboards. All changes you make in the UI and [CLI](./references/cli.md) are versionable because they are reflected as [Rill project files](./references/project-files.md) stored on disk. You can share your Rill project files with others by pushing to a shared repository, and they'll be able to completely recreate your project just by running `rill start`.

Have fun exploring Rill!

## We want to hear from you

You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Discord channel](https://bit.ly/3unvA05). Please abide by the [Rill Community Policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).
