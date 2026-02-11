---
title: Gemini
description: Use Gemini for AI features with your own API key
sidebar_label: Gemini
sidebar_position: 20
---

[Gemini](https://ai.google.dev/) is Google's family of AI models, designed for a wide range of reasoning, coding, and multimodal tasks. Rill supports connecting to Gemini using your own API key and configuration parameters to enable AI-powered conversations and data analysis features.

## API Key

Rill will use your configured Gemini API Connector if available, or fall back to our internal key if no custom configuration is provided. You can configure your API key in your project's `.env` file and reference the [credentials in a connector YAML](/reference/project-files/connectors#gemini).

### Gemini API Key

To configure Gemini access, you'll need to obtain an API key from Google AI Studio and configure it in your project.

1. **Obtain your Gemini API key:**
   - Sign in to [Google AI Studio](https://aistudio.google.com/)
   - Navigate to the [API Keys section](https://aistudio.google.com/apikey)
   - Create a new API key or use an existing one

2. **Configure the API key in your project:**

   For seamless deployment to Rill Cloud, you can configure this directly in your connector YAML:

   ```yaml
   type: connector
   driver: gemini
   api_key: "{{ .env.gemini_api_key }}"
   ```

:::tip Security Best Practice

Never commit your Gemini API key directly to your connector YAML files or version control. Always use environment variables with the `{{ .env.gemini_api_key }}` syntax to keep sensitive credentials secure.

:::

3. **Set up environment variable:**

   Configure the API key in your `.env` file:

   ```env
   gemini_api_key=AI...
   ```

4. **Configure Gemini as the default AI connector (optional):**

   If you want Gemini to be the default AI provider for your project, add the following to your `rill.yaml`:

   ```yaml
   ai_connector: gemini
   ```

You have now configured Gemini access for your Rill project. Rill will use these credentials to authenticate with Gemini when AI-powered features are utilized.

:::tip Cloud Credentials Management

If your project has already been deployed to Rill Cloud with configured credentials, you can use `rill env pull` to [retrieve and sync these cloud credentials](/developers/build/connectors/credentials/#rill-env-pull) to your local `.env` file. Note that this operation will overwrite any existing local credentials for this source.

:::

## Configuration Options

The Gemini connector supports additional configuration options for fine-tuning behavior:

| Property           | Description                                              | Example                       |
| ------------------ | -------------------------------------------------------- | ----------------------------- |
| `model`            | The Gemini model to use                                  | `gemini-2.5-pro-preview-05-06` |
| `max_output_tokens`| Maximum number of tokens in the response                 | `8192`                        |
| `temperature`      | Sampling temperature (0.0 - 2.0)                         | `0.0`                         |
| `top_p`            | Nucleus sampling parameter                               | `0.95`                        |
| `top_k`            | Top-K sampling parameter                                 | `40`                          |
| `include_thoughts` | Whether to include thinking/reasoning in the response    | `true`                        |
| `thinking_level`   | Level of thinking for the model's response               | `LOW`                         |

Example with all options:

```yaml
type: connector
driver: gemini
api_key: "{{ .env.gemini_api_key }}"
model: gemini-2.5-pro-preview-05-06
max_output_tokens: 8192
temperature: 0.0
```

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide Gemini API credentials used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#gemini) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:

```
rill env push
```

## Usage

Once configured, Gemini integration enables various AI-powered features in Rill:

- **Natural Language Queries**: Ask questions about your data using everyday conversational language on the dashboard or project level
- **Data Insights**: Get AI-generated insights and recommendations
- **Intelligent Suggestions**: Receive suggestions for dashboard improvements and data exploration

:::tip Don't see the AI Chat?

If you don't see AI-powered chat features in your Rill interface, [contact us](/contact) for the latest information on embedded chat capabilities and availability.

:::
