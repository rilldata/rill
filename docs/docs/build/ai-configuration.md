---
title: "AI Configuration"
description: "Configure AI instructions for your Rill project"
sidebar_label: "AI Configuration"
sidebar_position: 55
---

# AI Configuration

Rill's AI capabilities, including [AI Chat](/explore/ai-chat) and the [MCP Server](/explore/mcp), rely on context to provide accurate and relevant answers. You can provide additional context using the `ai_instructions` field in your project configuration files.

LLMs give their best results when they have good context. For a conversation with Rill Data, this means things like clarifying project-specific terms, routing questions to the correct metrics view, or defining business rules. Rather than expecting the user to provide this context every time, you can add `ai_instructions` to your project. This adds the context automatically for every conversation.

There are two places to add `ai_instructions`:

1.  **`rill.yaml`**: Project-wide instructions that apply to all queries across your entire project.
2.  **`<metrics_view>.yaml`**: Metrics view-specific instructions for individual dashboards.

## Automatic Context Inclusion

In addition to `ai_instructions`, Rill automatically includes the following in the AI context:

- **Measure and dimension descriptions**: Any `description` fields you add to measures and dimensions in your metrics view YAML files are automatically included in the AI context. This helps the AI understand what each metric or dimension represents without requiring you to duplicate that information in `ai_instructions`.
- **Metrics view metadata**: The metrics view name, display name, and description are included to help route questions to the correct dashboard.

This means you can document your measures and dimensions directly in your metrics view YAML, and that documentation will be available to the AI automatically.

## Project-Level Instructions ([`rill.yaml`](/build/project-configuration))

Use the `ai_instructions` field in `rill.yaml` to provide information that is **unique to your project**. This helps the AI agent deliver more relevant and actionable insights tailored to your specific needs.

**What to include:**
- Guidance on which metrics views are most important or should be prioritized for your project
- Any custom business logic, definitions, or terminology unique to your data or organization
- Preferences for aggregations, filters, or dimensions that are especially relevant to your use case
- Specific business context that helps the AI understand your domain

**Example:**

Here's an example of how you might configure `ai_instructions` in your `rill.yaml` to provide project context, metrics routing, and business definitions:

```yaml
ai_instructions: |
  # Project Context
  This project tracks e-commerce metrics for our multi-brand retail business.
  
  # Metrics View Routing
  - For questions about overall sales, revenue, or order volume → use `company_sales_metrics`
  - For questions about customer behavior, retention, or cohorts → use `customer_analytics`
  - For questions about product performance or inventory → use `product_metrics`
  - For questions about marketing campaigns or attribution → use `marketing_performance`
  - For questions about fulfillment, shipping, or logistics → use `operations_metrics`
  
  # Business Rules & Definitions
  - "Revenue" always refers to net revenue (after returns and discounts)
  - "Conversion rate" is calculated as orders/sessions, not users
  - Our fiscal year starts in February, not January
  - "Active customer" means a purchase within the last 90 days
  - Weekend traffic patterns are anomalous due to our B2B focus
  
  # Company Acronyms
  - GMV = Gross Merchandise Value
  - AOV = Average Order Value
  - ROAS = Return on Ad Spend
  - SKU = Stock Keeping Unit
  - NDR = Net Dollar Retention
  - CLTV = Customer Lifetime Value
  
  # Known Data Quirks
  - Mobile web data before March 2024 is incomplete due to tracking migration
  - European region data excludes VAT (use `revenue_with_vat` dimension if needed)
  - Refunds are processed with a 2-3 day delay, so recent data may shift
```

## Metrics View-Level Instructions ([`<metrics_view>.yaml`](/build/metrics-view/what-are-metrics-views))

You can provide context and instructions for AI tools interacting with a specific metrics view using the `ai_instructions` field in the metrics view's YAML file. This is useful for clarifying specific metrics, dimensions, or data quirks that apply only to that specific view.

:::tip Use descriptions for measure and dimension documentation
Instead of (or in addition to) adding definitions in `ai_instructions`, you can use the `description` key in the measure and dimension definitions in your metrics view YAML to document what each metric or dimension represents. These descriptions are automatically included in the AI context, making your metrics view self-documenting.
:::

**Example:**

```yaml
ai_instructions: |
  # Analysis Guidance
  - When analyzing "Revenue", always breakdown by "Region" to see currency impacts.
  - For questions about user growth, prioritize the "monthly_active_users" measure over "daily_active_users".
  - When comparing time periods, account for the fact that data for the "Legacy Plan" is static and will not update after Dec 2023.

  # Data Context
  - Mobile web data before March 2024 is incomplete due to tracking migration.
  - Refunds are processed with a 2-3 day delay, so recent data may shift.
  - Weekend traffic patterns are anomalous due to our B2B focus.
```

## Visualization Tips 

When using the [Rill MCP Server](/explore/mcp) with external AI clients like Claude, you can provide specific instructions on how to visualize data. Since the MCP server returns structured data, the AI client is responsible for rendering it.

:::note Visualization tips affect all AI interactions
Visualization instructions added to `rill.yaml` will affect both [Rill Chat](/explore/ai-chat) responses and external AI clients via the MCP Server. If you only want visualization tips to apply to external AI clients (like Claude Desktop), consider adding them to your client-specific configuration files instead:
- **Claude Desktop**: Add to `claude_desktop_config.json` or `Claude.md` in your project
- **Cursor**: Add to `.cursorrules` or `AGENT.md` in your project
- **Other AI clients**: Check your client's documentation for where to add custom instructions

This way, visualization formatting will only apply when using external clients, while Rill Chat maintains its default formatting.
:::

You can add instructions to your `rill.yaml` to guide the AI in presenting data more effectively (note that this will affect both Rill Chat and MCP clients):

```yaml
ai_instructions: |
  # Visualization Guidelines
  - When presenting time series data, use sparklines or trend indicators (e.g. 📈/📉) to show direction.
  - For comparisons, clearly state the percentage change and absolute difference.
  - Use bar charts for categorical comparisons when there are fewer than 10 categories.
  - When showing tables, always include a header row and align numeric columns to the right.
  
  # Example Formatting
  - Bar Charts using block characters:
    Q1 ████████░░ 411
    Q2 ██████████ 514
    Q3 ██████░░░░ 300
    Q4 ████████░░ 400

  - Horizontal progress bars: Project Progress:
    Frontend ▓▓▓▓▓▓▓▓░░ 80%
    Backend ▓▓▓▓▓▓░░░░ 60%
    Testing ▓▓░░░░░░░░ 20%
  
  - Using different block densities: Trends:
    Jan ▁▂▃▄▅▆▇█ High
    Feb ▁▂▃▄▅░░░ Medium
    Mar ▁▂░░░░░░ Low
    
  - Sparklines with Unicode Basic sparklines:
    Stock prices: ▁▂▃▅▂▇▆▃▅▇
    Website traffic: ▁▁▂▃▅▄▆▇▆▅▄▂▁
    CPU usage: ▂▄▆█▇▅▃▂▄▆█▇▄▂
    
  - Trend indicators: 
    AAPL ▲ +2.3% 
    GOOG ▼ -1.2% 
    MSFT ► +0.5% 
    TSLA ▼ -3.1%
  
  - Simple trend arrows: 
    Sales ↗️ (+15%)
    Costs ↘️ (-8%)
    Profit ⤴️ (+28%)
```
