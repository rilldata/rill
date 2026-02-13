---
title: Gemini
description: Use Gemini for AI features with your own API key
sidebar_label: Gemini
sidebar_position: 20
---

[Gemini](https://ai.google.dev/) is Google's family of AI models, designed for a wide range of reasoning, coding, and multimodal tasks. Rill supports connecting to Gemini using your own API key and configuration parameters to enable AI-powered conversations and data analysis features.

## API Key

Rill will use your configured Gemini connector if available, or fall back to its built-in LLM service if no custom configuration is provided. Once configured, your API key will be used for AI features in both Rill Developer and Rill Cloud, including the AI Agent builder. You can configure your API key in your project's `.env` file and reference the [credentials in a connector YAML](/reference/project-files/connectors#gemini).

### Gemini API Key

To configure Gemini access, you'll need to obtain an API key and configure it in your project.

1. **Obtain your Gemini API key** from [Google AI Studio](https://aistudio.google.com/apikey).

2. **Create the connector YAML:**

   Create `connectors/gemini.yaml` in your project:

   ```yaml
   type: connector
   driver: gemini
   api_key: "{{ .env.gemini_api_key }}"
   ```

3. **Set up environment variable:**

   If configuring manually, ensure your project's `.env` file contains the key before starting Rill:

   ```env
   gemini_api_key=AI...
   ```

4. **Configure Gemini as the default AI connector:**

   Add the following to your `rill.yaml` to use Gemini as the AI provider for your project:

   ```yaml
   ai_connector: gemini
   ```

For details on managing credentials across environments, see [Configure Local Credentials](/developers/build/connectors/credentials).

## Configuration Options

For additional configuration options (model, temperature, token limits, etc.), see the [Gemini connector reference](/reference/project-files/connectors#gemini).

## Deploy to Rill Cloud

Rill requires you to explicitly provide a Gemini API key to use the Gemini connector. See the [connector reference](/reference/project-files/connectors#gemini) for details.

For details on pushing and pulling credentials between environments, see [Configure Local Credentials](/developers/build/connectors/credentials#rill-env-push).
