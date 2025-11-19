---
title: "Analyze Your GitHub Repository"
sidebar_label: "GitHub Analytics"
sidebar_position: 20
hide_table_of_contents: false
tags:
  - Tutorial
  - Quickstart
  - Example Project
---

# Analyze Your GitHub Repository

This guide shows you how to build a powerful analytics dashboard for your own GitHub repository using Rill. In just a few commands, you'll have a fully interactive dashboard to discover hot zones in your codebase, track contributor activity, analyze code churn, and measure development velocity.

:::tip See it live
**[Explore the live demo →](https://ui.rilldata.com/demo/rill-github-analytics)** to see interactive dashboards analyzing real repositories (DuckDB, Rill). This is exactly what you'll create for your own repository.
:::

## Step 1: Clone the Project

### Clone from GitHub

```bash
# Clone the GitHub Analytics project
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-github-analytics
```

### Install Dependencies

The project includes Python scripts for downloading GitHub data and generating Rill project files. The project uses [Poetry](https://python-poetry.org/) for dependency management ([installation guide](https://python-poetry.org/docs/#installation)).

```bash
# Using Poetry (recommended)
poetry install

# Or using pip with venv
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install pandas pydriller
```

## Step 2: Scrape Git History and Save to Object Storage

**Prerequisites:**
- [Create a Google Cloud Storage bucket](https://cloud.google.com/storage/docs/creating-buckets)
- Set up GCS authentication:
  - [Create a service account key](https://cloud.google.com/iam/docs/keys-create-delete) with Storage Object Admin role
  - Set `GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json` environment variable

The project includes a `download_commits.py` script that clones your target repository, extracts commit metadata and file changes, then uploads the data as parquet files to your GCS bucket.

```bash
# Download and upload to GCS
python download_commits.py owner/repo --gcs --bucket gs://your-bucket/github-analytics

# Or limit to recent commits for faster testing
python download_commits.py owner/repo --gcs --bucket gs://your-bucket/github-analytics --limit 1000
```

**Note:** Files will be saved to `gs://your-bucket/github-analytics/owner/repo/` to keep data organized by repository.

:::note Private repositories
For private repos, use a fine-grained access token:

1. [Create a fine-grained personal access token](https://github.com/settings/tokens?type=beta) with **read-only** access to the repository
2. Store it as an environment variable: `export GITHUB_TOKEN=your_token_here`
3. Git will automatically use it when cloning (works in local dev and CI)
:::

**Note:** For large repositories with 10,000+ commits, the download may take 10-30 minutes. Use `--limit` to test with a smaller dataset first.

## Step 3: Generate Rill Project Files

The project includes a `generate_project.py` script that will generate:
- Source definitions pointing to your GCS bucket
- Data transformation models
- Metrics definitions
- An explore dashboard

Run the script with your repository and bucket:

```bash
# Generate Rill files configured for your GCS bucket
python generate_project.py owner/repo --gcs --bucket gs://your-bucket/github-analytics

# Examples:
python generate_project.py duckdb/duckdb --gcs --bucket gs://your-bucket/github-analytics
python generate_project.py your-org/your-repo --gcs --bucket gs://your-bucket/github-analytics
```

**Note:** Rill supports both Google Cloud Storage (GCS) and Amazon S3. The download script currently supports GCS. For S3, you'll need to modify the script.

:::note Just want to explore locally?
Use the `--local` flag instead: `python generate_project.py owner/repo --local`

This is great for testing, but you won't be able to deploy to Rill Cloud without migrating to cloud storage later.
:::

## Step 4: Deploy to Rill Cloud

Deploy your dashboard to share with your team:

```bash
rill deploy
```

This creates a live, shareable link where your team can explore the data together. Since your data is in cloud storage, Rill Cloud can access it directly.

:::tip Preview locally first
Want to verify everything looks good before deploying? Run `rill start` to preview the dashboard locally, then deploy when ready.
:::

## Next Steps

Now that you have GitHub Analytics deployed:

1. **Keep data fresh** – Schedule the download script to run regularly (cron, GitHub Actions, etc.) and keep your dashboards up to date
2. **Customize metrics** – Edit the metrics YAML files to add team-specific calculations
3. **Add alerts** – Use Rill's alerting features to monitor key metrics