---
title: Importing data
---

Rill ain't any fun without data! Here are instructions for getting data into Rill.

### Example data

If you want to give Rill a spin with example datasets, you can download some IoT and eCommerce data with this CLI command:

```bash
rill init-example
```

### Local file

Either use the Rill UI to select a local file from your computer, or see our docs on the [`rill import-source`](/cli#import-your-data) CLI command.

### AWS S3, Google Cloud Storage (GCS), and http(s)

In the Rill UI, navigate to a "Remote Source" connection and fill out the relevant form. A few things to note:
- Only parquet and csv files are supported.
- For S3 and GCS, you'll need to provide the region in which the bucket is located.
- Public buckets don't require any authentication, but private buckets require a service account key.
- You can import a single file or an entire bucket.

### DuckDB

See our docs on [importing data from an existing DuckDB database](/cli#existing-duckdb-databases).

### Request a new connector

If you don't see your data source listed above, [please let us know](https://discord.gg/eEvSYHdfWK)! We're continually adding new connectors, so your feedback will help us prioritize what data sources to support next.