---
title: Amazon Bedrock
description: Use models in your own AWS account for AI features
sidebar_label: Amazon Bedrock
sidebar_position: 5
---

[Amazon Bedrock](https://aws.amazon.com/bedrock/) is a managed AWS service for running foundation models. Rill can route its AI features (AI chat, generating metrics views and dashboards, and more) to a model in your own AWS account instead of Rill's built-in LLM service. Prompts, completions, and billing for model usage then stay within your AWS account.

## Requirements

- A Bedrock model that supports [tool use and structured outputs](https://docs.aws.amazon.com/bedrock/latest/userguide/conversation-inference-supported-models-features.html) in the Converse API. We recommend a recent Anthropic Claude model.
- [Model access](https://docs.aws.amazon.com/bedrock/latest/userguide/model-access.html) enabled for that model in your AWS account.
- Credentials with permission to call `bedrock:InvokeModel` on the model (or inference profile): either a Bedrock API key or AWS access keys.

## Setup

1. **Create the connector YAML:**

   Create `connectors/bedrock.yaml` in your project, using one of the authentication methods below:

   **Using a Bedrock API key** (quickest way to get started):

   ```yaml
   type: connector
   driver: bedrock

   model: us.anthropic.claude-sonnet-4-5-20250929-v1:0
   region: us-east-1
   api_key: "{{ .env.BEDROCK_API_KEY }}"
   ```

   Generate a long-term API key in the [Bedrock console](https://console.aws.amazon.com/bedrock) under **API keys**. A long-term key only works in the region it was generated in, so set `region` to match. Short-term API keys expire within 12 hours, so they are not suitable for a connector. Note that AWS [recommends long-term API keys for exploration only](https://docs.aws.amazon.com/bedrock/latest/userguide/api-keys.html); for production, use AWS access keys.

   **Using AWS access keys** (recommended for production):

   ```yaml
   type: connector
   driver: bedrock

   model: us.anthropic.claude-sonnet-4-5-20250929-v1:0
   region: us-east-1
   aws_access_key_id: "{{ .env.BEDROCK_AWS_ACCESS_KEY_ID }}"
   aws_secret_access_key: "{{ .env.BEDROCK_AWS_SECRET_ACCESS_KEY }}"
   ```

   The `region` is optional and defaults to `us-east-1`. When running Rill Developer locally, Rill uses the region from your local AWS configuration instead, if one is set.

   The `model` can be a foundation model ID, a cross-region inference profile ID (like the one above), or the ARN of an application inference profile, provisioned throughput, or imported custom model. Some models can only be called through an inference profile; if you see an error about on-demand throughput not being supported, use the model's inference profile ID instead (find it under **Cross-region inference** in the Bedrock console).

2. **Set up environment variables:**

   Ensure your project's `.env` file contains the credentials before starting Rill:

   ```env
   BEDROCK_API_KEY=...
   ```

   Or, for AWS access keys:

   ```env
   BEDROCK_AWS_ACCESS_KEY_ID=AKIA...
   BEDROCK_AWS_SECRET_ACCESS_KEY=...
   ```

   When running Rill Developer locally, you can omit the credentials and Rill will use the credentials from your local AWS configuration (for example, from `aws sso login` or the `AWS_BEARER_TOKEN_BEDROCK` environment variable).

3. **Configure Bedrock as the AI connector:**

   Add the following to your `rill.yaml` to route AI features to Bedrock:

   ```yaml
   ai_connector: bedrock
   ```

You can also add a Bedrock connector from the **Add AI connector** dialog in Rill Developer.

For details on managing credentials across environments, see [Configure Local Credentials](/developers/build/connectors/credentials).

## Assuming an IAM role

To call Bedrock using an IAM role (for example, in a different AWS account), set `aws_role_arn`. Rill uses the AWS access keys to assume the role before calling Bedrock:

```yaml
type: connector
driver: bedrock

model: us.anthropic.claude-sonnet-4-5-20250929-v1:0
region: us-east-1
aws_access_key_id: "{{ .env.BEDROCK_AWS_ACCESS_KEY_ID }}"
aws_secret_access_key: "{{ .env.BEDROCK_AWS_SECRET_ACCESS_KEY }}"
aws_role_arn: arn:aws:iam::123456789012:role/rill-bedrock
aws_external_id: "{{ .env.BEDROCK_AWS_EXTERNAL_ID }}" # Optional
```

Role assumption is not supported with API keys.

## Configuration Options

By default, Rill doesn't send `temperature` or `max_tokens`, so the model's defaults apply. You can set them in the connector, but leave `temperature` unset for models that don't support it, such as reasoning models. For all configuration options, see the [Bedrock connector reference](/reference/project-files/connectors#bedrock).

## Deploy to Rill Cloud

Rill Cloud does not have access to your local AWS configuration, so you must explicitly provide an API key or AWS access keys in the connector. Push your credentials to Rill Cloud with:

```bash
rill env push
```

For details on pushing and pulling credentials between environments, see [Configure Local Credentials](/developers/build/connectors/credentials#rill-env-push).
