---
title: Claude
description: Use Claude for AI features with your own API key
sidebar_label: Claude
sidebar_position: 10
---

[Claude](https://www.anthropic.com/claude) is Anthropic's AI assistant, designed to be helpful, harmless, and honest. Rill supports connecting to Claude using your own API key and configuration parameters to enable AI-powered conversations and data analysis features.

## API Key

Rill will use your configured Claude API Connector if available, or fall back to our internal key if no custom configuration is provided. You can configure your API key in your project's `.env` file and reference the [credentials in a connector YAML](/reference/project-files/connectors#claude).

### Claude API Key

To configure Claude access, you'll need to obtain an API key from your Anthropic account and configure it in your project.

1. **Obtain your Claude API key:**
   - Sign in to your [Anthropic Console](https://console.anthropic.com/)
   - Navigate to the [API Keys section](https://console.anthropic.com/settings/keys)
   - Create a new API key or use an existing one

2. **Configure the API key in your project:**

   For seamless deployment to Rill Cloud, you can configure this directly in your connector YAML:

   ```yaml
   type: connector
   driver: claude
   api_key: "{{ .env.CLAUDE_API_KEY }}"
   ```

:::tip Security Best Practice

Never commit your Claude API key directly to your connector YAML files or version control. Always use environment variables with the `{{ .env.CLAUDE_API_KEY }}` syntax to keep sensitive credentials secure.

:::

3. **Set up environment variable:**

   Configure the API key in your `.env` file:

   ```env
   claude_api_key=sk-ant-...
   ```

4. **Configure Claude as the default AI connector (optional):**

   If you want Claude to be the default AI provider for your project, add the following to your `rill.yaml`:

   ```yaml
   ai_connector: claude
   ```

You have now configured Claude access for your Rill project. Rill will use these credentials to authenticate with Claude when AI-powered features are utilized.

:::tip Cloud Credentials Management

If your project has already been deployed to Rill Cloud with configured credentials, you can use `rill env pull` to [retrieve and sync these cloud credentials](/developers/build/connectors/credentials/#rill-env-pull) to your local `.env` file. Note that this operation will overwrite any existing local credentials for this source.

:::

## Configuration Options

The Claude connector supports additional configuration options for fine-tuning behavior:

| Property      | Description                              | Example                                       |
| ------------- | ---------------------------------------- | --------------------------------------------- |
| `model`       | The Claude model to use                  | `claude-opus-4-5`, `claude-sonnet-4-20250514` |
| `max_tokens`  | Maximum number of tokens in the response | `8192`                                        |
| `temperature` | Sampling temperature (0.0 - 1.0)         | `0.0`                                         |
| `base_url`    | Custom base URL for the Claude API       | `https://api.anthropic.com`                   |

Example with all options:

```yaml
type: connector
driver: claude
api_key: "{{ .env.CLAUDE_API_KEY }}"
model: claude-sonnet-4-20250514
max_tokens: 8192
temperature: 0.0
```

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide Claude API credentials used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#claude) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:

```
rill env push
```

## Usage

Once configured, Claude integration enables various AI-powered features in Rill:

- **Natural Language Queries**: Ask questions about your data using everyday conversational language on the dashboard or project level
- **Data Insights**: Get AI-generated insights and recommendations
- **Intelligent Suggestions**: Receive suggestions for dashboard improvements and data exploration

:::tip Don't see the AI Chat?

If you don't see AI-powered chat features in your Rill interface, [contact us](/contact) for the latest information on embedded chat capabilities and availability.

:::
