---
title: "AI Chat in Rill Cloud"
description: Chat with your data using natural language in Rill Cloud
sidebar_label: "AI Chat"
sidebar_position: 00
---

## Overview

AI Chat in Rill Cloud allows you to have natural language conversations with your data directly in your browser. Instead of building queries or navigating through dashboards, simply ask questions using everyday conversational language and get instant insights backed by your metrics views—complete with **interactive charts and visualizations** that render right in the chat interface, plus **direct links** to your existing dashboards for deeper exploration.

AI Chat is powered by [Rill's Model Context Protocol (MCP)](/explore/mcp) integration, which ensures that responses are accurate, governed, and consistent with the metrics displayed in your dashboards. By querying data with **predefined measures and dimensions**, you can trust that the answers you receive are as reliable as the data in your Rill dashboards. 

**What makes AI Chat different?** Every response includes direct links to your Explore dashboards with filters pre-applied, so you can always **verify where the numbers came from**. No black box—just transparent, trustworthy analytics.

<img src='/img/explore/chat/project-chat.png' class='rounded-gif'/>
<br />

## How It Works 

AI Chat uses the same [Rill MCP Server](/explore/mcp) technology that powers external AI integrations with tools like Claude Desktop. This means:

- **Fast!** - Get instant answers powered by Rill's optimized query engine 
- **Accurate Responses** - The Agent only queries [metrics views](/build/metrics-view) you've already defined, ensuring accuracy and consistency
- **Secure Data Access** - Respects your [project's access](/build/metrics-view/security) controls and user permissions

## Accessing AI Chat

### Access AI Chat from Project Home

1. Navigate to your [Rill Cloud](https://ui.rilldata.com) project home page
2. Click on the **AI** tab in the project navigation
3. Start typing your question in the chat interface

<!-- 
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
::: -->

## Understanding Responses

AI Chat provides rich, multi-layered responses to help you understand your data quickly while maintaining easy access to deeper exploration:

### What's Included in Responses

1. **Summary** - A concise answer to your question with key findings and insights
2. **Interactive Visualizations** - Charts and graphs that help you see patterns at a glance. The AI automatically chooses the most appropriate visualization based on your data and question, including:
   - **Line charts** - Show trends and changes over time
   - **Area charts** - Highlight cumulative trends and patterns
   - **Bar charts** - Compare values across categories or dimensions
   - **Stacked bar charts** - Show part-to-whole relationships across categories
   - **Donut charts** - Display proportional breakdowns of a total
   - **Combo charts** - Combine multiple measures with different scales
   - **Heatmaps** - Visualize distribution across two dimensions

3. **Detailed Results** - Tables or lists with specific numbers and breakdowns
4. **Dashboard Links** - Direct links to your existing [Explore dashboards](/explore/dashboard-101) with filters and time ranges pre-applied based on your question
5. **Suggested Follow-ups** - Related questions or areas to investigate further

### Visual Components

AI Chat automatically generates interactive visualizations to complement answers when appropriate. The AI intelligently selects chart types based on your data structure, question context, and visualization best practices—while skipping charts when a table or text-only response is more suitable.

### Linking Back to Dashboards

**Trust the numbers.** The most powerful feature of AI Chat is its transparency. When the AI answers your question, it automatically generates links to your Explore dashboards, allowing you to **see exactly where the numbers came from**. 

Every AI response includes dashboard links that:

- **Pre-apply relevant filters** - The dashboard opens with filters matching your question context
- **Set appropriate time ranges** - Time periods from your question are automatically selected
- **Enable comparison periods** - When you ask about changes, comparison views are activated

This means you can **verify every answer** by clicking through to the full dashboard. No black box—just transparent, trustworthy analytics. Start with a quick AI summary, then explore the underlying data with full confidence in its accuracy.


## Improving AI Chat with Instructions

To get the most accurate and contextual responses from AI Chat, you can add custom `ai_instructions` to your project files. These instructions provide the AI with additional context about your data, business logic, and preferred response formats.

### Why Add AI Instructions?

AI instructions help the AI:
- Understand your specific business context and terminology
- Format responses in ways that match your team's preferences
- Focus on the metrics and dimensions most relevant to your use case

### Where to Add Instructions

You can add `ai_instructions` in two places:

1. **`rill.yaml`** - Project-wide instructions that apply to all queries across your entire project
2. **`<metrics_view>.yaml`** - Metrics view-specific instructions for individual dashboards

For detailed examples and best practices on writing effective AI instructions, see the [Rill MCP documentation](/explore/mcp#adding-ai-instructions-to-your-model).

## Use in Your Favorite AI Client

Prefer to chat with your data in Claude Desktop, ChatGPT, or another AI assistant? You can connect your Rill projects to external AI clients using the **[Rill MCP Server](/explore/mcp)**. This gives you the same governed, accurate analytics experience—powered by your predefined metrics—but integrated into your preferred AI workflow. Perfect for data teams who want deep analysis sessions, local development access, or integration with other tools. See the **[Rill MCP Server documentation](/explore/mcp)** to learn more.

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
