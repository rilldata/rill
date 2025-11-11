---
title: HTTPS
description: Connect to remote data sources via HTTP/HTTPS
sidebar_label: HTTPS
sidebar_position: 25
---

Import data from remote sources accessible via HTTP or HTTPS URLs into your Rill project.

## Overview

The HTTPS connector allows you to import data from publicly accessible files hosted on web servers, CDNs, or cloud storage services. This is perfect for working with datasets that are regularly updated or shared publicly.


## Connect to HTTPS

To connect to data via HTTPS, you have two options depending on your data source:

1. **Use authenticated endpoints** (for private APIs or protected resources)
2. **Use public URLs** (for publicly accessible files)

Choose the method that matches your data source requirements.

### Authenticated Endpoints

If your endpoint requires authentication, create a `https` connector that can test your API key before importing data:

```yaml
type: connector 
driver: https 
path: "https://api.endpoint.com/v1" 

headers:
    Authorization: "Bearer {{ .env.connector.https.token }}"
```

### Public URLs

For publicly accessible files, you can simply use the `https` connector to add data from your endpoint or publicly available file.

## Adding an HTTPS Source

### Option 1: Using the Rill UI

1. In the left navigation pane, click the **"Add"** button and select data
2. Select **"HTTPS"** from the connector options
3. Enter your file's URL (e.g., `https://example.com/data.csv`)
4. Click **"Add Data"**




### Option 2: Using Code

Create a YAML configuration file in your project's `sources` directory:

```yaml
type: model
connector: https

path: https://example.com/data.csv
```

## Supported URL Types

The HTTPS connector supports various remote data sources:

- **Public HTTP/HTTPS URLs**: Direct links to CSV, JSON, Parquet, and other data files
- **Cloud Storage URLs**: Public links to files in Google Cloud Storage, Amazon S3, or other cloud providers

## Authentication

### Local Development

When running Rill locally, you can use existing credentials configured on your machine for authenticated requests. If you don't have credentials configured, you'll need to define the key with `headers: "Authorization: Bearer token"`.

### Deploy to Rill Cloud

When deploying to Rill Cloud, you must explicitly provide service account credentials with appropriate access permissions for protected resources.

## Best Practices

- **Use HTTPS URLs**: Prefer secure HTTPS connections over HTTP when possible
- **Check File Accessibility**: Ensure your URLs are publicly accessible or properly authenticated
- **Monitor File Size**: Large files may take longer to download and process
- **Regular Updates**: Set up automated refreshes for frequently updated datasets

## Reference

For advanced configuration options and properties, see the [Connector YAML Reference](/reference/project-files/connectors#https).