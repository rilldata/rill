# Rill MCP Desktop Extension

Connect Claude Desktop to your Rill metrics views for governed data analytics.

## What is this?

This Desktop Extension packages the Rill MCP server for easy installation in Claude Desktop. It uses the proven `mcp-remote` package to connect to your existing Rill runtime endpoints. Instead of manually editing configuration files, you can install this extension with one click and configure it through a simple GUI.

## Features

- **List metrics views** - Discover available metrics views in your Rill project
- **Get specifications** - Understand dimensions and measures for each metrics view
- **Query time ranges** - Find available data time spans
- **Execute queries** - Run aggregation queries with filters, sorting, and limits

## Supported Connections

- **Local Rill Developer** - Connect to `localhost:9009` for local development
- **Public Rill Cloud Projects** - Connect to publicly accessible projects
- **Private Rill Cloud Projects** - Connect using personal access tokens

## Installation

1. Download the `.dxt` file
2. Open Claude Desktop
3. Go to Settings → Extensions
4. Drag and drop the `.dxt` file
5. Click "Install"
6. Configure your connection settings

## Configuration

### Local Development

- Select "Local Rill Developer"
- Ensure Rill is running on `localhost:9009`

### Cloud Projects

- Select "Public" or "Private" Rill Cloud Project
- Enter your Organization ID and Project ID (from the project URL)
- For private projects, create a personal access token: `rill token issue`

## Usage

Once installed, you can ask Claude questions like:

- "What metrics views are available in my project?"
- "Show me the revenue trends for the last quarter"
- "What are the top performing campaigns by region?"
- "Compare week-over-week growth in user signups"

Claude will use the Rill MCP tools to query your data and provide accurate, governed analytics results.

## Documentation

For more information, visit: https://docs.rilldata.com/explore/mcp
