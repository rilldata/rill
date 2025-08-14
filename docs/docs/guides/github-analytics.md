---
title: "GitHub Analytics Demo"
sidebar_label: "GitHub Analytics Demo"
sidebar_position: 5
hide_table_of_contents: false
tags:
  - Tutorial
  - Quickstart
  - Example Project
---

# GitHub Analytics Demo

This guide walks you through the GitHub Analytics demo project, which showcases Rill's capabilities for analyzing GitHub repository data. You'll learn how to clone the project, understand its structure, and explore the analytics dashboard.

## Overview

The GitHub Analytics demo analyzes data from the ClickHouse repository, providing insights into:
- **Commit activity** – Daily commits, authors, and patterns
- **File changes** – Most modified files, line additions/deletions
- **Contributor analysis** – Top contributors and their activity
- **Repository trends** – Growth patterns and development velocity

## Step 1: Clone the Project

### Clone from GitHub

```bash
# Clone the GitHub Analytics project
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-github-analytics
```

### Alternative: Use Rill CLI

```bash
# Clone directly using Rill CLI, assuming you have access to the demo project.
rill org switch demo
rill project clone rill-github-analytics
cd rill-github-analytics
```

## Step 2: Project Structure

The project is organized as follows:

```
rill-github-analytics/
├── rill.yaml                           # Project configuration
├── sources/                            # Data source definitions
│   ├── commits.yaml                    # GitHub commits data
│   └── modified_files.yaml             # File modification data
├── models/                             # SQL transformations
│   └── XXX_commits_model.sql           # Where XXX is the repository
├── metrics/                            # Defined measures and dimensions
│   └── XXX_commits_metrics.yaml        # Where XXX is the repository
├── dashboards/                         # Dashboard configurations
│   ├── repo_compare_canvas.yaml        # Canvas dashboard
│   └── rill_commits_explore.yaml       # Explore dashboard
└── README.md                           # Project documentation
```

## Step 3: Data Sources

### Commits Source

The commits source connects to Google Cloud Storage to fetch GitHub commit data:

```yaml
# sources/rill_commits_source.yaml
type: source
connector: "gcs"
uri: "gs://rilldata-public/github-analytics/rilldata/rill/commits/commits*.parquet"
```

**What this does:**
- Connects to the `rilldata-public` GCS bucket
- Fetches commit data from the Rill Data repository
- Uses glob patterns to get data across multiple years/months
- Data includes commit hashes, authors, dates, and messages

### Modified Files Source

The modified files source tracks file changes:

```yaml
# sources/rill_modified_files.yaml
type: source
connector: "gcs"
uri: "gs://rilldata-public/github-analytics/rilldata/rill/commits/modified_files*.parquet"
```

**What this does:**
- Tracks which files were modified in each commit
- Records line additions and deletions
- Enables analysis of code churn and file popularity

## Step 4: Data Models

### Joining the sources
Without going into a full deep dive, this involves joining the commit details and modified files sources based on the commit hash.

```sql
-- Model SQL
-- Reference documentation: https://docs.rilldata.com/build/models
-- @materialize: true

SELECT
    author_date AS date,
    c.commit_hash,
    commit_msg AS commit_message,
    author_name AS username,
    merge AS is_merge_commit,
    new_path AS file_path,
    filename,
    RIGHT(filename, POSITION('.' IN REVERSE(filename))) AS file_extension,
    CASE WHEN CONTAINS(file_path, '/')
      THEN SPLIT_PART(file_path, '/', 1)
      ELSE NULL
    END AS first_directory,
    CASE WHEN CONTAINS(SUBSTRING(file_path, LENGTH(first_directory) + 2), '/')
      THEN SPLIT_PART(file_path, '/', 2)
      ELSE NULL
    END AS second_directory,
    CASE 
      WHEN first_directory IS NOT NULL AND second_directory IS NOT NULL
        THEN CONCAT(first_directory, '/', second_directory) 
      WHEN first_directory IS NOT NULL
        THEN first_directory
      WHEN first_directory IS NULL
        THEN NULL
    END AS second_directory_concat,
    added_lines AS additions,
    deleted_lines AS deletions, 
    additions + deletions AS changes, 
    old_path AS previous_file_path,
FROM rill_commits_source c
LEFT JOIN rill_modified_files f ON c.commit_hash = f.commit_hash
```

## Step 5: Creating your Metrics View

Metrics in Rill define the measures and dimensions that power your dashboards. Let's create a metrics file to define the key analytics for GitHub commits:

```yaml
# metrics/rill_commits_metrics.yaml
title: "Rill Commits Metrics"
description: "Key metrics for analyzing Rill repository activity"

measures:
  - name: "total_commits"
    description: "Total number of commits"
    expression: "COUNT(*)"
    
  - name: "unique_contributors"
    description: "Number of unique contributors"
    expression: "COUNT(DISTINCT username)"
    
  - name: "total_changes"
    description: "Total lines of code changed (additions + deletions)"
    expression: "SUM(changes)"
    
  - name: "total_additions"
    description: "Total lines of code added"
    expression: "SUM(additions)"
    
  - name: "total_deletions"
    description: "Total lines of code deleted"
    expression: "SUM(deletions)"
    
  - name: "avg_commit_size"
    description: "Average number of changes per commit"
    expression: "AVG(changes)"
    
  - name: "merge_commit_ratio"
    description: "Percentage of commits that are merge commits"
    expression: "AVG(CASE WHEN is_merge_commit THEN 1 ELSE 0 END)"

dimensions:
  - name: "date"
    description: "Commit date"
    expression: "date"
    
  - name: "username"
    description: "GitHub username"
    expression: "username"
    
  - name: "file_extension"
    description: "File extension"
    expression: "file_extension"
    
  - name: "first_directory"
    description: "First directory in file path"
    expression: "first_directory"
    
  - name: "second_directory_concat"
    description: "Combined first and second directories"
    expression: "second_directory_concat"
    
  - name: "commit_message"
    description: "Commit message text"
    expression: "commit_message"

time_grain: "day"
```

**What this metrics file does:**

- **Measures** define the calculations you want to perform:
  - `total_commits` – Count of all commits
  - `unique_contributors` – Number of different developers
  - `total_changes` – Lines of code modified
  - `avg_commit_size` – Average complexity of commits
  - `merge_commit_ratio` – Percentage of merge commits

- **Dimensions** define how you can slice and dice the data:
  - `date` – Time-based analysis
  - `username` – Per-contributor analysis
  - `file_extension` – Analysis by file type
  - `first_directory` – Analysis by code area

- **Time grain** sets the default time aggregation to daily.

### Creating Custom Metrics

You can add more sophisticated metrics:

```yaml
# Additional measures for advanced analysis
measures:
  - name: "active_days"
    description: "Number of days with commits"
    expression: "COUNT(DISTINCT date)"
    
  - name: "commit_frequency"
    description: "Commits per active day"
    expression: "total_commits / active_days"
    
  - name: "code_churn_ratio"
    description: "Ratio of deletions to additions"
    expression: "CASE WHEN total_additions > 0 THEN total_deletions / total_additions ELSE 0 END"
```

## Step 6: Dashboard Exploration

#### **Features**
- **Explore Slice-and-Dice** – For data exploration and ad-hoc analysis
- **Canvas** – Traditional charts and visualizations
- **Pivot/Flat Table** – Tabular data views with sorting and grouping
- **Measure's TDD** – Granular analysis of a single measure

#### **Selectors**
- **Date range selector** - Analyze specific time periods
- **Time Comparison Toggle** - Compare previous time periods
- **Dimensions Comparison** - Compare unique dimension values

#### **Filters**
- **Author filter** - Focus on specific contributors
- **File type filter** - Analyze specific file types


### Let's answer some questions basic questions!

What parts of your codebase are the most active? What parts have the most churn?
<img src='/img/tutorials/quickstart/github-analytics-1.png' class='rounded-gif'/>
<br />

How large are commits? What do commits that touch many files have in common?
<img src='/img/tutorials/quickstart/github-analytics-2.png' class='rounded-gif'/>
<br />

How productive are your contributors? How does productivity change week over week?
<img src='/img/tutorials/quickstart/github-analytics-3.png' class='rounded-gif'/>
<br />


These are just some of the insights that you can find within your explore dashboard but you'll find more hidden gems in your data as you continue to use Rill. Please let us know if you have any other questions!