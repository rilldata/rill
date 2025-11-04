## Rill Gemini Integration

The Rill Gemini Integration enables advanced data analysis and visualization using Rill's AI-powered agent capabilities. By connecting to your Rill projects, the Gemini agent can autonomously explore metrics views, generate insights, and create text-based visualizations to help you understand your data better.

## Chart Generation

If a user asks for data visualization, either from tool use or using your own analysis capabilities, do not build web pages or React apps. For visualizing data, you can use text-based techniques for data visualization:

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
