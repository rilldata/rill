# Gemini CLI Extension for Rill Data

Professional data analysis and report generation using Rill's metrics layer with Google Docs integration.

## Features

- **Comprehensive Analytics**: Discover, analyze, and visualize metrics data
- **Professional Reports**: Generate executive-ready Google Docs with automated formatting
- **Local & Cloud Storage**: Create reports locally first, then optionally upload to Google Drive
- **Stakeholder Sharing**: Automatically share reports with configurable permissions

Check out the [Rill Data website](https://www.rilldata.com) for more information about Rill.

Example report generated using this extension can be found in the [Gemini Documentation](docs/bids_report.md).

## Installation

Install the extension via GitHub:

```bash
gemini extensions install https://github.com/rilldata/gemini
```

Install Rill CLI if you haven't already:

```bash
curl https://rill.sh | sh
```

## Configuration

Generate a Rill authentication token:

```bash
rill token issue --display-name "Gemini Extension"
```

Update `gemini-extension.json` with your Rill credentials:

- `RILL_PROJECT_URL`: Your Rill project's MCP endpoint
- `RILL_AUTH_HEADER`: Your authentication token

For Google Drive integration, ensure you have:

- Valid Google Cloud credentials configured (see [GEMINI.md](GEMINI.md))
- Appropriate scopes: `drive`, `documents`

## Usage

### Quick Example

> Generate a change report comparing sales metrics from last month to this month using the `sales_dashboard` metrics view

This extension will:

1. Query your Rill metrics for both time periods
2. Analyze the differences in key metrics
3. Create a formatted Google Doc with the comparison
4. Share it and return the link

### Basic Analysis

```bash
gemini rill analyze "Show me user engagement trends for the last quarter"
```

### Report Generation

#### Create a Report

```bash
gemini rill generate "Create a monthly performance report"
```

#### Create Local Report, Then Upload Later

```bash
# First, create locally
gemini rill generate "Create quarterly analysis"

# Review the report, then upload when ready
gemini rill upload_local_report --filePath "./reports/quarterly-analysis-2025-10-03.md"
```

## Available Tools

- **`generate`** - Create comprehensive analytics reports
  - Required: `title`, `content`
  - Options: `saveLocalOnly`, `localFilePath`, `shareEmail`
- **`export_to_sheet`** - Export Rill query results to Google Sheets
  - Required: `title`, `data`
  - Options: `sheetName`, `shareEmail`

> **Note**: Additional tools like `create_metrics_comparison_report`, `update_report`, `upload_local_report`, and `list_reports` are planned for future releases.

## Documentation

For detailed documentation, visit the [Rill Docs](https://docs.rilldata.com/), and refer to the [Local Development Guide](LOCAL.md) for setting up a development environment.
