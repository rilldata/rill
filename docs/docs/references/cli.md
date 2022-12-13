---
title: CLI
description: Start and manage Rill using the command-line interface (CLI)
sidebar_label: CLI
sidebar_position: 40
---

## Start a new project

In any empty folder, simply run:

```bash
rill start
```

to initialize an empty project and open the Rill web app on [http://localhost:9009](http://localhost:9009). You can use the web app to define data sources, data models and dashboards.

## Help menu

To see usage information about all the available CLI commands, run:

```bash
rill
```

For any sub-command, you can always pass `--help` to output detailed usage information. For example:

```bash
rill start --help
```

outputs detailed information about the flags you can pass to `rill start`.

## Initializing an empty project

In any empty folder, run

```bash
rill init
```

to initialize an empty project.

## Initializing an example project

We recommend following our [quick start](../quick-start.md) to experience how well Rill ties together with Git. However, Rill also bundles some built-in examples to help you get started quickly. To initialize the default example, run:

```
rill init --example
```

To view a list of all built-in example projects:

```
rill init --list-examples
```

To use a non-default example, pass it as a parameter to `--example`:

```
rill init --example=sf_props
```

## Running Rill in another directory

You can explicity specify a project folder outside of the current folder using the `--project` option:

```
rill init --project /path/to/project
rill source add /path/to/data.parquet --project /path/to/project
rill start --project /path/to/project
```

## Import a local data file

You can create a local file source by running:

```
rill source add /path/to/data.parquet
```

See [Import data](../using-rill/import-data.md) for more details.

### Override the source name

By default the source name will be a sanitized version of the dataset file name. You can specify a name using the `name` command.

```
rill source add /path/to/data.parquet --name my_source
```

### Custom CSV delimiters

If you have a CSV file that is delimited by a character other than a comma or tab, you can use the `--delimiter` option. If a delimiter is not set, Rill automatically tries to detect the delimiter, so this is not strictly necessary:

```
rill source add /path/to/data.csv --delimiter "|"
```

## Dropping a source

If you have added a source to Rill that you want to drop, run:

```bash
rill source drop my_source
```
