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

This guide shows you how to build a powerful analytics dashboard for **your own GitHub repository** using Rill. In just a few commands, you'll have a fully interactive dashboard analyzing commits, contributors, file changes, and development patterns.

:::tip See it live
**[Explore the live demo →](https://ui.rilldata.com/demo/rill-github-analytics)** to see interactive dashboards analyzing real repositories (DuckDB, Rill). This is exactly what you'll create for your own repository.
:::

<img src='/img/tutorials/quickstart/github-analytics-1.png' class='rounded-gif'/>

## What you'll discover

- **Hot zones** – Which files and directories change most
- **Contributor activity** – Who's building what and how much
- **Code churn** – Addition vs. deletion ratios over time
- **Development velocity** – Commit frequency and trends

## Step 1: Clone the Project

### Clone from GitHub

```bash
# Clone the GitHub Analytics project
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-github-analytics
```

### Install Dependencies

The project includes Python scripts for downloading GitHub data. The project uses [Poetry](https://python-poetry.org/) for dependency management ([installation guide](https://python-poetry.org/docs/#installation)).

```bash
# Using Poetry (recommended)
poetry install

# Or using pip with venv
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install pandas pydriller
```

## Step 2: Generate Rill Project Files

Create Rill project files for your GitHub repository:

```bash
# Generate Rill files for any GitHub repository
python setup_repo.py owner/repo

# Examples:
python setup_repo.py duckdb/duckdb
python setup_repo.py your-org/your-repo
```

This creates:
- Source definitions pointing to cloud storage
- Data transformation models
- Metrics definitions
- An explore dashboard

**Note:** Rill supports both Google Cloud Storage (GCS) and Amazon S3. The download script currently supports GCS. For S3, you'll need to modify the script.

:::note Just want to explore locally?
Add the `--local` flag to use local files: `python setup_repo.py owner/repo --local`

This is great for testing, but you won't be able to deploy to Rill Cloud without migrating to cloud storage later.
:::

## Step 3: Scrape Git History and Save to Object Storage

**Prerequisites:**
- [Create a Google Cloud Storage bucket](https://cloud.google.com/storage/docs/creating-buckets)
- Configure GCS credentials: run `gcloud auth login` or set up a [service account key](https://cloud.google.com/iam/docs/keys-create-delete)

Extract commit history and save to GCS:

```bash
# Download and upload to GCS
python download_commits.py owner/repo --gcs --bucket gs://your-bucket/github-analytics

# Or limit to recent commits for faster testing
python download_commits.py owner/repo --gcs --bucket gs://your-bucket/github-analytics --limit 1000
```

The script will:
- Clone the repository
- Extract commit metadata and file changes
- Upload data as parquet files to your GCS bucket

**Note:** For large repositories with 10,000+ commits, the download may take 10-30 minutes. Use `--limit` to test with a smaller dataset first.

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