# Gemini CLI Extension for Rill Data

Professional data analysis and report generation using Rill's metrics layer.

## Overview

Check out [Rill's Docs](https://docs.rilldata.com) for more information about Rill.

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

## Development

### Local Development

To test changes locally:

1. Make your changes to the extension files
2. Install the extension from your local development branch:
   ```bash
   gemini extensions install https://github.com/rilldata/rill --ref=your-branch-name
   ```

### Releasing

Extension updates are automatically available when changes are merged to the main branch. Users will be prompted to update when new versions are available.

The extension version is managed through the `gemini-extension.json` file, and users can install specific versions using git refs if needed.
