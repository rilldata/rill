---
title: OpenAI
description: Connect to OpenAI services for AI-powered features
sidebar_label: OpenAI
sidebar_position: 30
---

[OpenAI](https://openai.com/) provides powerful AI services including language models, embeddings, and other machine learning capabilities. Rill supports connecting to OpenAI services using your own API key and configuration parameters to enable AI-powered conversations and data analysis features.

## API Key

Rill will use your configured OpenAI connector if available, or fall back to its built-in LLM service if no custom configuration is provided. Once configured, your API key will be used for AI features in both Rill Developer and Rill Cloud, including the AI Agent builder. You can configure your API key in your project's `.env` file and reference the [credentials in a connector YAML](/reference/project-files/connectors#openai).

### OpenAI API Key

To configure OpenAI access, you'll need to obtain an API key and configure it in your project.

1. **Obtain your OpenAI API key** from the [OpenAI Platform](https://platform.openai.com/api-keys).

2. **Create the connector YAML:**

   Create `connectors/openai.yaml` in your project:

   ```yaml
   type: connector
   driver: openai
   api_key: "{{ .env.OPENAI_API_KEY }}"
   ```

3. **Set up environment variable:**

   If configuring manually, ensure your project's `.env` file contains the key before starting Rill:

   ```env
   OPENAI_API_KEY=sk-...
   ```
   
4. **Configure OpenAI as the default AI connector:**

   Add the following to your `rill.yaml` to use OpenAI as the AI provider for your project:

   ```yaml
   ai_connector: openai
   ```
For details on managing credentials across environments, see [Configure Local Credentials](/developers/build/connectors/credentials).

## Configuration Options

For additional configuration options (model, base URL, API type, etc.), see the [OpenAI connector reference](/reference/project-files/connectors#openai).

### OpenAI-compatible APIs

The connector can also target APIs that implement the OpenAI chat completions protocol. Configure the provider's URL and model on the connector. If the provider supports JSON mode but not OpenAI's JSON Schema response format, set `structured_output_mode` to `json_object`:

```yaml
type: connector
driver: openai
api_key: "{{ .env.PROVIDER_API_KEY }}"
base_url: https://llm.example.com/v1
model: example-model
structured_output_mode: json_object
```

For provider-specific request extensions, use the advanced `extra_body` map. Its values must be JSON-serializable, and it cannot override core request fields such as `model`, `messages`, `tools`, `response_format`, or streaming controls. For example, an OpenAI-compatible endpoint can receive a chat-template option as follows:

```yaml
extra_body:
  chat_template_kwargs:
    enable_thinking: false
```

These options belong to the connector, so different connectors in the same Rill process can use different provider behavior.

## Deploy to Rill Cloud

Rill requires you to explicitly provide an OpenAI API key to use the OpenAI connector. See the [connector reference](/reference/project-files/connectors#openai) for details.

For details on pushing and pulling credentials between environments, see [Configure Local Credentials](/developers/build/connectors/credentials#rill-env-push).
