---
title: AI Chat
sidebar_label: AI Chat
sidebar_position: 70
---

## Overview

The AI chat feature in Rill allows you to ask questions about your data in natural language. The AI analyzes your question, queries your data, and provides answers with inline visualizations when appropriate. This makes data exploration more intuitive and accessible, without requiring knowledge of SQL or query syntax.

:::info

The AI chat feature is available in Rill Cloud only. It is not available in Rill Developer.

:::

## How it works

When you ask a question, the AI:

1. Analyzes your question and the structure of your data
2. Generates and executes the appropriate SQL query
3. Returns results with a natural language explanation
4. Automatically creates visualizations when they help communicate the answer

The AI has access to your project's metrics views and can answer questions about:

- Trends and patterns in your data
- Comparisons between different segments
- Aggregations and calculations
- Time-based analysis

## Visualizations in AI chat

The AI automatically generates inline chart visualizations to help answer your questions. Charts appear directly in the chat interface alongside text explanations.

### Available chart types

The AI can create the following chart types:

- **Line charts** – For time series data and trends
- **Area charts** – For showing cumulative trends over time
- **Bar charts** – For comparing values across categories
- **Stacked bar charts** – For comparing multiple series across categories
- **Donut charts** – For showing proportions of a whole
- **Combo charts** – For displaying multiple measures with different scales
- **Heatmaps** – For showing patterns across two dimensions

### When charts are generated

The AI generates charts after running queries, unless the results are:

- Single values or simple metrics
- Primarily text-based data
- Tables where visualization wouldn't add clarity
- Explicitly excluded by your question phrasing

The AI intelligently selects the most appropriate chart type based on your data structure and the context of your question.

:::note

To enable AI chat with visualizations, you need to enable the `chatCharts` feature flag in your `rill.yaml` file. See [Feature flags](/build/project-configuration#feature-flags) for more information.

:::

## Using AI chat

To use AI chat:

1. Navigate to a dashboard in Rill Cloud
2. Click the chat icon in the top right corner
3. Type your question in natural language
4. Press Enter or click Send

The AI will process your question and return an answer with visualizations if applicable.

### Example questions

Here are some example questions you can ask:

- "What were the top 5 products by revenue last month?"
- "Show me the trend in daily active users over the past quarter"
- "How does conversion rate compare across different regions?"
- "What's the average order value for customers who purchased more than once?"

:::tip

For best results, be specific in your questions. Include time periods, metrics, and dimensions you're interested in.

:::

## Privacy and security

- AI chat queries only have access to data in your Rill project
- Queries respect the same security policies and row-level security as the rest of Rill
- Chat history is stored securely and associated with your user account
- Your data is not used to train AI models

:::info

For more information about Rill's security practices, see our [Security documentation](/deploy/security).

:::
