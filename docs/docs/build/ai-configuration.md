---
title: "AI Configuration"
description: "Configure AI instructions for your Rill project"
sidebar_label: "AI Configuration"
sidebar_position: 55
---

# AI Configuration

Rill's AI capabilities, including [AI Chat](/explore/ai-chat) and the [MCP Server](/explore/mcp), rely on context to provide accurate and relevant answers. You can provide this context using the `ai_instructions` field in your project configuration files.

LLMs give their best results when they have good context. For a conversation with Rill Data, this means things like clarifying project-specific terms, routing questions to the correct metrics view, or defining business rules. Rather than expecting the user to provide this context every time, you can add `ai_instructions` to your project. This adds the context automatically for every conversation.

There are two places to add `ai_instructions`:

1.  **`rill.yaml`**: Project-wide instructions that apply to all queries across your entire project.
2.  **`<metrics_view>.yaml`**: Metrics view-specific instructions for individual dashboards.

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

**Example:**

```yaml
ai_instructions: |
  # Measure Definitions
  - "Churn Rate" excludes trial users who cancelled within 7 days.
  - "Active Users" are defined as users with at least one login in the selected period.

  # Data Context
  - Data for the "Legacy Plan" is static and will not update after Dec 2023.
  - When analyzing "Revenue", always breakdown by "Region" to see currency impacts.
```

## Visualization Tips 

When using the [Rill MCP Server](/explore/mcp) with external AI clients like Claude, you can provide specific instructions on how to visualize data. Since the MCP server returns structured data, the AI client is responsible for rendering it.

You can add instructions to your `rill.yaml` to guide the AI in presenting data more effectively:

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
