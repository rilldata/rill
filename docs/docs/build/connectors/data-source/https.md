---
title: HTTPS
description: Connect to remote data sources via HTTP/HTTPS
sidebar_label: HTTPS
sidebar_position: 25
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

The HTTPS connector allows you to import data from remote sources accessible via HTTP or HTTPS URLs into your Rill project. It supports publicly accessible files hosted on web servers, CDNs, or cloud storage services, making it perfect for working with datasets that are regularly updated or shared publicly.

The connector supports various remote data sources:

- **Public HTTP/HTTPS URLs**: Direct links to CSV, JSON, Parquet, and other data files hosted on web servers or CDNs
- **Cloud Storage URLs**: Public links to files in Google Cloud Storage, Amazon S3, Azure Blob Storage, or other cloud providers
- **REST API Endpoints**: Connect to REST APIs that return JSON, CSV, or other structured data formats. The connector can handle paginated responses and authenticated endpoints

The connector supports downloading and processing various file formats:

- **CSV**: Comma-separated values files
- **JSON**: JavaScript Object Notation files (including JSONL/NDJSON)
- **Parquet**: Columnar data format
- **Other formats**: Any format supported by DuckDB's file readers

## Using Public URLs

For publicly accessible files, you don't need to create a connector. Simply use the URL directly in your source or model configuration.

### Using the UI

1. In the left navigation pane, click the **"Add"** button and select data
2. Select **"HTTPS"** from the connector options
3. Enter your file's URL (e.g., `https://example.com/data.csv`)
4. Click **"Add Data"**

### Using Code

Create a model that reads directly from an HTTPS URL:

```yaml
type: model
materialize: true

connector: duckdb

sql: |
  select * from read_csv('https://example.com/data.csv', auto_detect=true, ignore_errors=1, header=true)
```

For REST API endpoints that return JSON:

```yaml
type: model
materialize: true

connector: duckdb

sql: |
  select * from read_json('https://api.example.com/data.json', auto_detect=true)
```

---

## Using Authenticated Endpoints

If your endpoint requires authentication, you need to follow a two-step process:

1. **Create a connector** - Configure your HTTPS connector with authentication credentials
2. **Create a source model** - Define which URL to ingest using the connector

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

### Using the UI

When you add an HTTPS data model through the Rill UI for authenticated endpoints, the process follows two steps:

1. **Configure Authentication** - Set up your HTTPS connector with authentication headers
2. **Configure Data Model** - Define which URL to ingest

The UI will automatically create both the connector file and model file for you.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_https.yaml`:

```yaml
type: connector 
driver: https 

headers:
    Authorization: "Bearer {{ .env.connector.https.token }}"
```

**Step 2: Create model configuration**

Create `models/my_https_data.yaml`:

```yaml
type: model
materialize: true

connector: duckdb

sql: |
  select * from read_csv('https://api.example.com/data.csv', auto_detect=true, ignore_errors=1, header=true)
```

**Step 3: Add credentials to `.env`**

```bash
connector.https.token=your_api_token_here
```

:::note
For advanced configuration options and properties, see the [Connector YAML Reference](/reference/project-files/connectors#https).
:::

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide authentication credentials for protected HTTPS endpoints used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#https) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```


## Best Practices

- **Use HTTPS URLs**: Prefer secure HTTPS connections over HTTP when possible to protect data in transit
- **Check File Accessibility**: Ensure your URLs are publicly accessible or properly authenticated before deployment
- **Monitor File Size**: Large files may take longer to download and process. Consider using incremental models for large datasets
- **Regular Updates**: Set up automated refreshes for frequently updated datasets using [model refresh configuration](/build/models/data-refresh)
- **Error Handling**: Use `ignore_errors=1` in SQL functions when reading files to handle malformed rows gracefully
- **Rate Limiting**: Be aware of API rate limits when connecting to REST API endpoints. Consider implementing retry logic or pagination handling

