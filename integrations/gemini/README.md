# Gemini CLI Extension for Rill Data

Professional data analysis and report generation using Rill's metrics layer.

## Overview

Check out [Rill's Docs](https://docs.rilldata.com) for more information about Rill.

## Installation

Install the extension via GitHub using a specific release tag:

```bash
gemini extensions install https://github.com/rilldata/rill --ref=gemini-v1.0.0
```

> **Note**: Since this is a monorepo, you must specify a Gemini-specific release tag (prefixed with `gemini-v`) to ensure Gemini CLI picks up the correct extension files. See [Releases](https://github.com/rilldata/rill/releases) for available versions.

For the latest version, check the [releases page](https://github.com/rilldata/rill/releases) and use the most recent `gemini-v*` tag.

Install Rill CLI if you haven't already:

```bash
curl https://rill.sh | sh
```

## Configuration

After installation, configure the extension with your Rill credentials. The extension will prompt you for the following information during setup:

- **Organization**: Your Rill organization name
- **Project**: Your Rill project name
- **Access Token**: Your Rill access token

To generate a Rill authentication token, run:

```bash
rill token issue --display-name "Gemini Extension"
```

> **Tip**: You can find your organization and project names in the Rill Cloud UI URL: `https://ui.rilldata.com/{organization}/{project}`

The extension configuration is handled automatically through Gemini's settings interface - no manual environment file setup is required.

## Development

### Local Development

To test changes locally:

1. Make your changes to the extension files
2. Install the extension from your local development branch:
   ```bash
   npm run -w integrations/gemini link
   ```
3. Test the extension in Gemini:
   ```bash
   npm run -w integrations/gemini unlink
   ```

### Releasing

This extension is distributed through [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
