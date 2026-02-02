---
title: Claude
description: Use Claude for AI features with your own API key
sidebar_label: Claude
sidebar_position: 43
---

Rill supports connecting to Claude using your own API key and configuration parameters for AI-powered coding and data analysis features.

## Steps

1. Obtain an API key from the [Claude Console](https://platform.claude.com).
2. Add the API key as an environment variable in your Rill project called `claude_api_key`.
    - See [Configure Credentials](/developers/build/connectors/credentials) for details.
    - Remember to sync the updated environment variables with `rill env pull` or `rill env push`.
3. Create a Claude connector YAML file in your Rill project:
   ```yaml
   # connectors/claude.yaml
   type: connector
   driver: claude
   api_key: "{{ .env.claude_api_key }}"
   model: "claude-opus-4-5"
   ```
4. Configure `claude` as the project's default AI connector in `rill.yaml`:
   ```yaml
   # rill.yaml
   # ...
   ai_connector: claude
   ```
