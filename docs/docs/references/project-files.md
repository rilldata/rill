---
title: Rill project files
sidebar_label: Rill project files
sidebar_position: 50
---
When you create sources, models, and metrics versions we create code-representations on your behalf on the file system. You can see these files in your `source`, `models` and `dashboards` folders in your project. 

Projects can be "re-hydrated" from Rill project files into an explorable data application - figuring out the dependencies, pulling down data, & validating your model queries and metrics configurations. The result is a set of functioning exploratory dashboards.

You can see an example project by visiting our [example github repository](https://github.com/rilldata/rill-developer-example.git).


## Source connections
In your Rill project directory, create a `source.yaml` file in the `sources` directory containing a `type` and location (`uri` or `path`). Rill will automatically detect and ingest the source next time you run `rill start`.

**`type`**
 —  the type of connector you are using for the source _(required)_. Possible values include:
  - _`https`_ — public files available on the web.
  - _`s3`_ — a file available on amazon s3.
  - _`gcs`_ — a file available on google cloud platform.
  - _`local_file`_ — a locally available file.

**`uri`**
 —  the URI of the remote connector you are using for the source _(required for type: http, s3, gcs)_.
 Additionally Rill also supports glob patterns as part of URI including recursive searches for s3 and gcs.
  - _`s3://your-org/bucket/file.parquet`_ —  the s3 URI of your file
  - _`gs://your-org/bucket/file.parquet`_ —  the gsutil URI of your file
  - _`https://data.example.org/path/to/file.parquet`_ —  the web address of your file

**`path`**
 — the _local path_ of the connector you are using for the source relative to your project's root directory.   _(required for type: file)_
- _`/path/to/file.csv`_ —  the path to your file

**`region`**
 — Optionally sets the cloud region of the bucket you want to connect to. Only available for S3.
  - _`us-east-1`_ —  the cloud region identifer

**`glob.max_total_size`**
 — Appplicable if the URI is a glob pattern. The max allowed total size(in bytes) of all objects matching the glob pattern.
  - default value is _`10737418240 (10GB)`_

**`glob.max_objects_matched`**
 — Appplicable if the URI is a glob pattern. The max allowed number of objects matching the glob pattern.
  - default value is _`10,000`_

**`glob.max_objects_listed`**
 — Appplicable if the URI is a glob pattern. The max number of objects to list and match against glob pattern (excluding files excluded by the glob prefix).

  - default value is _`100,000`_

See our Using Rill guide for an [example](../using-rill/import-data#using-code).

## Model transformation
Data transformations in Rill Developer are powered by DuckDB and their dialect of SQL (duckSQL). Under the hood, data models are created as views in DuckDB. Please visit their [documentation](https://duckdb.org/docs/sql/introduction) for insight into how to write your queries.

In your Rill project directory, create a `model_name.sql` file in the `models` directory containing a duckSQL `SELECT` statement. Rill will automatically detect and parse the model next time you run `rill start`.

## Dashboard metrics

In your Rill project directory, create a `dashboard_name.yaml` file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.


_**`model`**_ — the model name powering the dashboard with no path _(required)_

_**`display_name`**_ — the display name for the dashboard _(required)_

_**`timeseries`**_ — column from your model that will underlie x-axis data in the line charts _(required)_

_**`dimensions:`**_ — for exploring [segments](../using-rill/metrics-dashboard#dimensions) and filtering the dashboard _(required)_
  - _**`property`**_ — a categorical column _(required)_ 
  - _**`label`**_ — a label for your dashboard dimension _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 

_**`measures:`**_ — numeric [aggregates](../using-rill/metrics-dashboard#measures) of columns from your data model  _(required)_
  - _**`expression`**_ — a combination of operators and functions for aggregations _(required)_ 
  - _**`label`**_ — a label for your dashboard measure _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 
  - _**`format_preset`**_ — one of a set of values that format dashboard measures. _(optional; default is humanize)_. Possible values include:
      - _`humanize`_ — round off numbers in an opinionated way to thousands (K), millions (M), billions B), etc
      - _`none`_ — raw output
      - _`currency_usd`_ —  output rounded to 2 decimal points prepended with a dollar sign
      - _`percentage`_ — output transformed from a rate to a percentage appended with a percentage sign
      - _`comma_separators`_ — output transformed to decimal formal with commas every 3 digits

See our Using Rill guide for an [example](../using-rill/metrics-dashboard#using-code).

## Project definition

- _**`name`**_ — the name of your project which will be displayed in the upper left hand corner
- _**`compiler`**_ — the version of the runtime compiler that is compatible with the artifacts (ex: `rill-beta`)
- _**`rill_version`**_ — the version of Rill Developer  that is compatible with the artifacts (ex: `v0.16`)
