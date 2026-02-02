export const EMPTY_PROJECT_TITLE = "Untitled Rill Project";

export const EXAMPLES = [
  {
    name: "rill-cost-monitoring",
    title: "Cost Monitoring",
    description: "Monitoring cloud infrastructure",
    image: "/img/welcome-bg-cost-monitoring.png",
    firstFile: "/dashboards/margin_scorecard.yaml",
    connector: "duckdb",
  },
  {
    name: "rill-openrtb-prog-ads",
    title: "OpenRTB Programmatic Ads",
    description: "Real-time Bidding (RTB) advertising",
    image: "/img/welcome-bg-openrtb.png",
    firstFile: "/dashboards/auction_explore.yaml",
    connector: "duckdb",
  },
  {
    name: "rill-github-analytics",
    title: "Github Analytics",
    description: "A Git project's commit activity",
    image: "/img/welcome-bg-github-analytics.png",
    firstFile: "/dashboards/clickhouse_commits_explore.yaml",
    connector: "clickhouse",
  },
];
