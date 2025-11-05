# Gemini CLI Extension for Rill Data

Professional data analysis and report generation using Rill's metrics layer.

## Overview

Check out [Rill's Docs](https://docs.rilldata.com) for more information about Rill.

## Installation

Install the extension via GitHub:

```bash
gemini extensions install https://github.com/rilldata/rill
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
   npm run -w integrations/gemini link
   ```
3. Test the extension in Gemini:
   ```bash
   npm run -w integrations/gemini unlink
   ```

### Releasing

This extension is distributed through [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases). This provides a faster and more reliable initial installation experience for users, as releases are shipped as single archives instead of requiring a git clone.

Each release includes an archive file containing the full contents of the repository at the tagged commit. When checking for updates, Gemini CLI looks for the "latest" release on GitHub (you must mark it as such when creating the release), unless the user installed a specific release by passing `--ref=<some-release-tag>`.

The extension version is managed through the `gemini-extension.json` file, and releases are automated via GitHub Actions when new version tags are pushed.
