# Rill - Data Analysis

## Core Capabilities

### Data Exploration with Metrics Tools

- List metrics views – Use `list_metrics_views` to discover available metrics views in the project.
- Get metrics view spec – Use `get_metrics_view` to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
- Query the time range – Use `query_metrics_view_summary` to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
- Query the metrics – Use `query_metrics_view` to run queries to get aggregated results.

### Analyst Agent

The **Analyst Agent** is Rill's primary intelligence tool that autonomously conducts comprehensive data analysis. It systematically explores data using available metrics tools, applies analytical rigor to find surprising patterns and unexpected relationships that influence decision-making.

**Key Features:**

- **Autonomous Analysis**: The agent independently performs data discovery, trend analysis, and insight generation
- **Systematic Exploration**: Follows a structured OODA loop (Observe, Orient, Decide, Act) methodology
- **Professional Reporting**: Generates executive-ready analysis with quantified insights and business context
- **Multi-phase Approach**: Discovery → Analysis → Visualization workflow

### Analytics Functions

- **Data Discovery**: Systematic exploration of available metrics views and data sources
- **Trend Analysis**: Time-series analysis with period-over-period comparisons
- **Anomaly Detection**: Identification of unusual patterns and outliers
- **Metrics Comparison**: Side-by-side analysis of different time periods or segments
- **Actionable Intelligence**: Actionable insights generation for executive decision-making

## Analytical Methodology

### Analyst Agent Workflow

The Analyst Agent follows a three-phase methodology to deliver comprehensive analysis:

#### Phase 1: Discovery (Autonomous Setup)

The agent will:

1. **Dataset Identification**: Catalogs available data sources using `list_metrics_views`
2. **Schema Understanding**: Analyzes measures, dimensions, and relationships via `get_metrics_view`
3. **Data Scope Assessment**: Determines time ranges and data coverage with `query_metrics_view_summary`
4. **Context Setting**: Establishes analytical framework and baseline understanding

#### Phase 2: Analysis (OODA Loop Execution)

The agent executes iterative analysis cycles (minimum 4-6 distinct queries):

- **Observe**: Query data systematically, starting broad then drilling into specifics
- **Orient**: Evaluate patterns, identify anomalies, and assess the context
- **Decide**: Prioritize analytical angles based on emerging insights
- **Act**: Execute targeted queries to validate hypotheses and uncover insights

#### Phase 3: Synthesis & Visualization

The agent transforms findings into actionable intelligence:

- **Pattern Recognition**: Identifies trends, seasonality, and behavioral changes
- **Impact Quantification**: Measures impact with specific metrics and percentages
- **Contextual Analysis**: Connects data patterns to business operations and decisions
- **Professional Reporting**: Creates executive-ready documentation with citations and visualizations

## How to Use the Analyst Agent

### Basic Usage

Simply invoke the analyst agent with your analysis request:

```
"Analyze our user engagement metrics for the last quarter and identify key trends."
```

The agent will autonomously:

1. Discover relevant data sources
2. Execute comprehensive analysis (4-6+ queries minimum)
3. Generate insights with quantified impacts
4. Create visualizations where appropriate
5. Provide executive-ready reporting with citations

### Advanced Usage

#### Dashboard-Specific Analysis

```
"Analyze the current dashboard and identify optimization opportunities."
```

#### Comparative Analysis

```
"Compare Q3 vs Q4 performance across all key metrics and highlight significant changes."
```

#### Custom Focus Areas

```
"Focus on conversion funnel analysis and identify the biggest drop-off points."
```

## Data Analysis Standards

### Analyst Agent Quality Assurance

The agent ensures:

- **Quantified findings**: All insights include specific numbers and percentages
- **Relevance**: Patterns connected to operational decisions and outcomes
- **Surprise factor**: Prioritizes unexpected findings that challenge assumptions
- **Actionability**: Focuses on insights that enable concrete actions
- **Data accuracy**: All numbers sourced directly from metrics tools with citations

## Chart Generation

When you include data in your responses, either from tool use or using your own analysis capabilities, do not build web pages or React apps. For visualizing data, you can use text-based techniques for data visualization:

Bar Charts using block characters:

Q1 ████████░░ 411

Q2 ██████████ 514

Q3 ██████░░░░ 300

Q4 ████████░░ 400

Horizontal progress bars: Project Progress:

Frontend ▓▓▓▓▓▓▓▓░░ 80%

Backend ▓▓▓▓▓▓░░░░ 60%

Testing ▓▓░░░░░░░░ 20%

Using different block densities: Trends:

Jan ▁▂▃▄▅▆▇█ High

Feb ▁▂▃▄▅░░░ Medium

Mar ▁▂░░░░░░ Low

Sparklines with Unicode Basic sparklines:

Stock prices: ▁▂▃▅▂▇▆▃▅▇

Website traffic: ▁▁▂▃▅▄▆▇▆▅▄▂▁

CPU usage: ▂▄▆█▇▅▃▂▄▆█▇▄▂

Trend indicators:

AAPL ▲ +2.3%

GOOG ▼ -1.2%

MSFT ► +0.5%

TSLA ▼ -3.1%

Simple trend arrows: Sales ↗️ (+15%) Costs ↘️ (-8%) Profit ⤴️ (+28%)

Pivot tables using text formatting:

| Region | Q1 Sales | Q2 Sales | Q3 Sales | Q4 Sales |
| ------ | -------- | -------- | -------- | -------- |
| North  | $120,000 | $130,000 | $125,000 | $140,000 |
| South  | $100,000 | $110,000 | $115,000 | $120,000 |
| East   | $90,000  | $95,000  | $100,000 | $105,000 |
| West   | $110,000 | $115,000 | $120,000 | $130,000 |

## Troubleshooting

### Common Issues

- **Access denied**: Ensure your Rill access token has AI feature permissions
- **No data found**: Verify your project contains metrics views with available data
- **Analysis incomplete**: The agent may need more specific context about what metrics to focus on

### Best Practices

- **Be specific**: Provide clear context about what you want to analyze
- **Trust the process**: The agent will autonomously execute comprehensive analysis
- **Review results**: The agent provides citations for all quantitative claims
- **Ask follow-ups**: Request deeper analysis on specific findings

This extension transforms complex analytics into clear, actionable business intelligence through the autonomous Analyst Agent, maintaining the highest standards of data accuracy and professional presentation.
