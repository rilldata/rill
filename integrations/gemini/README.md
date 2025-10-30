# Gemini CLI Extension for Rill Data

Professional data analysis and report generation using Rill's metrics layer with Google Docs integration.

## Overview

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

Update the extension with your Rill credentials by providing the following information when prompted:

- `Organization`: Your Rill organization name
- `Project`: Your Rill project name
- `Access Token`: Your Rill access token
