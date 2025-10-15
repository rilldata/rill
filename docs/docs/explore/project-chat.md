---
title: "AI Chat in Rill Cloud"
description: Chat with your data using natural language in Rill Cloud
sidebar_label: "AI Chat"
sidebar_position: 05
---

## Overview

AI Chat in Rill Cloud allows you to have natural language conversations with your data directly in your browser. Instead of building queries or navigating through dashboards, simply ask questions in plain English and get instant insights backed by your metrics views—complete with **interactive visualizations** and **direct links** to your existing dashboards for deeper exploration.

AI Chat is powered by [Rill's Model Context Protocol (MCP)](/explore/mcp) integration, which ensures that responses are accurate, governed, and consistent with the metrics displayed in your dashboards. By querying data with **predefined measures and dimensions**, you can trust that the answers you receive are as reliable as the data in your Rill dashboards. Plus, responses include canvas dashboard components that visualize your data and automatically link back to the relevant dashboards with filters pre-applied.

<img src='/img/explore/chat/project-chat.png' class='rounded-gif'/>
<br />

## How It Works

AI Chat uses the same [Rill MCP Server](/explore/mcp) technology that powers external AI integrations with tools like Claude Desktop. This means:

- **Governed Data Access** - The AI only queries [metrics views](/build/metrics-view) you've already defined, ensuring accuracy and consistency
- **Structured Responses** - Results are based on predefined measures and dimensions, not raw database queries
- **Context-Aware** - The AI understands your metrics view structure, including available dimensions, measures, and time ranges
- **Secure** - Respects your [project's access](/build/metrics-view/security) controls and user permissions

## Accessing AI Chat

### Access AI Chat from Project Home

1. Navigate to your [Rill Cloud](https://ui.rilldata.com) project home page
2. Click on the **AI** tab in the project navigation
3. Start typing your question in the chat interface

### Access AI Chat from a Dashboard

You can also access AI Chat directly while exploring a dashboard, making it easy to ask questions about what you're currently viewing:

1. While viewing any [Explore dashboard](/explore/dashboard-101), look for the **AI Chat icon** in the top navigation bar
2. Click the AI Chat icon to open the chat panel alongside your dashboard
3. Ask questions about the data you're currently viewing
<img src='/img/explore/chat/dashboard-chat.png' class='rounded-gif'/>
<br />


When you open AI Chat from a dashboard, the AI is automatically aware of:
- **Current dashboard context** - The metrics view you're viewing
- **Applied filters** - Any dimension or measure filters you've set
- **Time range** - The time period currently selected
- **Comparison settings** - Any active time comparisons

This context-aware functionality means you can ask questions like:
- "Why did this metric spike?" (referring to what's visible on screen)
- "What's driving this change?" (analyzing the current time period)
- "Show me more details about these results" (diving deeper into filtered data)

:::tip Context-Aware Queries
Opening AI Chat from within a dashboard allows for more natural, context-aware questions. The AI understands what you're looking at, so you don't need to repeat filters or time ranges in your questions.
:::

## Understanding Responses

AI Chat provides rich, multi-layered responses to help you understand your data quickly while maintaining easy access to deeper exploration:

### What's Included in Responses

1. **Summary** - A concise answer to your question with key findings and insights
2. **Canvas Dashboard Components** - Interactive visualizations built using canvas dashboard widgets that help you see patterns at a glance. These can include:
   - Time series charts showing trends over time
   - Bar charts comparing dimensions or categories
   - Tables with formatted data
   - Big number displays for key metrics
3. **Detailed Results** - Tables or lists with specific numbers and breakdowns
4. **Dashboard Links** - Direct links to your existing [Explore dashboards](/explore/dashboard-101) with filters and time ranges pre-applied based on your question
5. **Suggested Next Steps** - Follow-up questions or areas to investigate further

### Linking Back to Dashboards

One of the most powerful features of AI Chat is how it seamlessly connects back to your existing Rill dashboards. When the AI answers your question, it automatically generates links that:

- **Pre-apply relevant filters** - The dashboard opens with filters matching your question context
- **Set appropriate time ranges** - Time periods from your question are automatically selected
- **Select relevant metrics** - The measures and dimensions discussed in the chat are highlighted
- **Enable comparison periods** - When you ask about changes, comparison views are activated

This means you can **start with a quick AI answer**, then **click through to the full dashboard** for deeper, interactive exploration - all without manually setting up filters or searching for the right view.

### Visual Components

Unlike text-only AI assistants, AI Chat in Rill Cloud can render actual dashboard components directly in the chat interface. These visualizations are built using the same [canvas dashboard](/build/dashboards/canvas-widgets) technology used throughout Rill, ensuring:

- **Consistency** - Visualizations match the style and behavior of your regular dashboards
- **Interactivity** - Click, hover, and interact with charts right in the chat
- **Accuracy** - Charts are generated from the same data sources as your dashboards
- **Clarity** - Complex data patterns become immediately visible

:::tip From Chat to Dashboard
Think of AI Chat as your starting point for data exploration. Use it to quickly answer questions and spot trends, then leverage the dashboard links to dive deeper with full filtering, comparison, and export capabilities.
:::

## Improving AI Chat with Instructions

To get the most accurate and contextual responses from AI Chat, you can add custom `ai_instructions` to your project files. These instructions provide the AI with additional context about your data, business logic, and preferred response formats.

### Why Add AI Instructions?

AI instructions help the AI:
- Understand your specific business context and terminology
- Format responses in ways that match your team's preferences
- Generate properly formatted Explore dashboard links
- Focus on the metrics and dimensions most relevant to your use case
- Provide more actionable insights tailored to your workflows

### Where to Add Instructions

You can add `ai_instructions` in two places:

1. **`rill.yaml`** - Project-wide instructions that apply to all queries across your entire project
2. **`<metrics_view>.yaml`** - Metrics view-specific instructions for individual dashboards

For detailed examples and best practices on writing effective AI instructions, see the [Rill MCP documentation](/explore/mcp#adding-ai-instructions-to-your-model).


## AI Chat vs. Rill MCP Server

Rill offers two ways to chat with your data using AI:

| Feature      | AI Chat (in Rill Cloud)                    | Rill MCP Server                               |
| ------------ | ------------------------------------------ | --------------------------------------------- |
| **Location** | Built into Rill Cloud browser interface    | External AI assistants (Claude Desktop, etc.) |
| **Setup**    | No setup required                          | Requires MCP client configuration             |
| **Access**   | Any Rill Cloud user with project access    | Requires personal access token                |
| **Use Case** | Quick questions while exploring dashboards | Deep analysis sessions in your AI assistant   |
| **Best For** | Business users and analysts                | Data teams and power users                    |

Both use the same underlying technology and provide equally accurate results based on your metrics views.

## Best Practices
<!-- 
### Start Broad, Then Narrow
Begin with general questions to understand the data, then ask follow-up questions to dive deeper:

1. "What are my top performing products?"
2. "Show me the revenue trend for Product X over the last quarter"
3. "Which regions drive the most revenue for Product X?"

### Use Follow-Up Questions
The AI maintains context within a conversation, so you can ask follow-up questions without repeating information:

- Initial: "What was total revenue last month?"
- Follow-up: "How does that compare to the previous month?"
- Follow-up: "Which product categories drove the increase?"

### Leverage Explore Links
When the AI provides an Explore link, click through to the dashboard for:
- Interactive filtering and drilling down
- Applying additional comparisons
- Creating bookmarks or scheduled reports
- Exporting data

### Combine with Dashboards
Use AI Chat for quick answers and discovery, then switch to [interactive dashboards](/explore/dashboard-101) when you need:
- Fine-grained control over filters
- Multiple simultaneous comparisons
- Visual exploration of dimension relationships
- Creating alerts or scheduled reports -->

## Need Help?

[Contact our team](/contact) if you have questions or feedback about AI Chat!