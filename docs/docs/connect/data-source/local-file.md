---
title: Local File
description: Connect to your local data files
sidebar_label: Local File
sidebar_position: 35
---

Import data from files stored on your local machine into your Rill project.

## Overview

The Local File connector allows you to import data from CSV, JSON, Parquet, and other supported file formats directly from your local filesystem. This is perfect for working with local datasets, exports, or files you've downloaded.

## Adding a Local File Source

### Option 1: Using the Rill UI

1. In the left navigation pane, click the **"+"** button next to **Sources**
2. Select **"Local File"** from the connector options
3. Navigate to and select your file, or drag and drop it directly onto the Rill interface

### Option 2: Using Code

Create a YAML configuration file in your project's `sources` directory:

```yaml
type: source
connector: local_file
path: /path/to/your/data.csv
```

**Important:** When using relative paths, they should be relative to your Rill project root (where `rill.yaml` is located), not the `sources` directory.

## Importing Multiple Files

Use glob patterns to import data from multiple files at once:

```yaml
type: source
connector: local_file
path: /path/to/data/*.csv
```

**Examples:**
- `data/*.csv` - All CSV files in the data directory
- `exports/2024-*.parquet` - All Parquet files from 2024
- `logs/**/*.json` - All JSON files in logs and subdirectories

For detailed glob pattern syntax, refer to the [DuckDB multiple files documentation](https://duckdb.org/docs/stable/data/multiple_files/overview.html).

## Supported File Formats

The Local File connector supports various file formats including:
- CSV
- JSON
- Parquet


:::warning File Size Limits
When ingesting the Data into Rill, you'll notice a new `/data` folder path with a copy of the CSV file. This is designed so that when you publish to Rill Cloud, the file will also be included. Note that there is a 100MB limit to each unique file. Files over 100MB will not be deployed with your project.
:::
