---
title: OpenAI
description: Connect to OpenAI services for AI-powered features
sidebar_label: OpenAI
sidebar_position: 42
---


[OpenAI](https://openai.com/) provides powerful AI services including language models, embeddings, and other machine learning capabilities. Rill supports connecting to OpenAI services using your own API key and configuration parameters to enable AI-powered conversations and data analysis features. 

## API Key

Rill will use your configured OpenAI API Connector if available, or fall back to our internal key if no custom configuration is provided. You can configure your API key in your project's `.env` file and referencing the [credentials in a connector YAML](/reference/project-files/connectors#openai).

### OpenAI API Key

To configure OpenAI access, you'll need to obtain an API key from your OpenAI account and configure it in your project.

1. **Obtain your OpenAI API key:**
   - Sign in to your [OpenAI account](https://platform.openai.com/)
   - Navigate to the [API Keys section](https://platform.openai.com/api-keys)
   - Create a new API key or use an existing one

2. **Configure the API key in your project:**

   For seamless deployment to Rill Cloud, you can configure this directly in your connector YAML:

   ```yaml
   type: connector
   driver: openai
   api_key: "{{ .env.connector.openai.openai_api_key }}"
   ```
:::tip Security Best Practice

Never commit your OpenAI API key directly to your connector YAML files or version control. Always use environment variables with the `{{ .env.connector.openai.openai_api_key }}` syntax to keep sensitive credentials secure.

:::
3. **Set up environment variable:**
   
   Configure the API key in your `.env` file:

   ```env
   connector.openai.openai_api_key=sk-...
   ```

You have now configured OpenAI access for your Rill project. Rill will use these credentials to authenticate with OpenAI services when AI-powered features are utilized.

:::tip Cloud Credentials Management

If your project has already been deployed to Rill Cloud with configured credentials, you can use `rill env pull` to [retrieve and sync these cloud credentials](/build/connectors/credentials/#rill-env-pull) to your local `.env` file. Note that this operation will overwrite any existing local credentials for this source.

:::

## Usage

Once configured, OpenAI integration enables various AI-powered features in Rill:

- **Natural Language Queries**: Ask questions about your data using everyday conversational language on the dashboard or project level.
- **Data Insights**: Get AI-generated insights and recommendations
- **Intelligent Suggestions**: Receive suggestions for dashboard improvements and data exploration

:::tip Don't see the AI Chat?

If you don't see AI-powered chat features in your Rill interface, [contact us](/contact) for the latest information on embedded chat capabilities and availability.

:::

