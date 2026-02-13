---
title: Claude
description: Use Claude for AI features with your own API key
sidebar_label: Claude
sidebar_position: 10
---

[Claude](https://www.anthropic.com/claude) is Anthropic's AI assistant, designed to be helpful, harmless, and honest. Rill supports connecting to Claude using your own API key and configuration parameters to enable AI-powered conversations and data analysis features.

## API Key

Rill will use your configured Claude connector if available, or fall back to its built-in LLM service if no custom configuration is provided. You can configure your API key in your project's `.env` file and reference the [credentials in a connector YAML](/reference/project-files/connectors#claude).

### Claude API Key

To configure Claude access, you'll need to obtain an API key and configure it in your project.

1. **Obtain your Claude API key** from the [Anthropic Console](https://console.anthropic.com/settings/keys).

2. **Create the connector YAML:**

   Create `connectors/claude.yaml` in your project:

   ```yaml
   type: connector
   driver: claude
   api_key: "{{ .env.claude_api_key }}"
   ```

3. **Set up environment variable:**

   If configuring manually, ensure your project's `.env` file contains the key before starting Rill:

   ```env
   claude_api_key=sk-ant-...
   ```

4. **Configure Claude as the default AI connector:**

   Add the following to your `rill.yaml` to use Claude as the AI provider for your project:

   ```yaml
   ai_connector: claude
   ```

For details on managing credentials across environments, see [Configure Local Credentials](/developers/build/connectors/credentials).

## Configuration Options

For additional configuration options (model, temperature, token limits, etc.), see the [Claude connector reference](/reference/project-files/connectors#claude).

## Deploy to Rill Cloud

Rill requires you to explicitly provide a Claude API key to use the Claude connector. See the [connector reference](/reference/project-files/connectors#claude) for details.

For details on pushing and pulling credentials between environments, see [Configure Local Credentials](/developers/build/connectors/credentials#rill-env-push).
