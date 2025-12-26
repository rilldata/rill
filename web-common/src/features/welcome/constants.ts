import DuckDBIcon from "@rilldata/web-common/components/icons/connectors/DuckDBIcon.svelte";
import ClickHouseIcon from "@rilldata/web-common/components/icons/connectors/ClickHouseIcon.svelte";

export const EMPTY_PROJECT_TITLE = "Untitled Rill Project";

export const EXAMPLES = [
  {
    name: "rill-cost-monitoring",
    title: "Cost Monitoring",
    description: "Monitoring cloud infrastructure",
    image: "/img/welcome-bg-cost-monitoring.png",
    firstFile: "margin_scorecard.yaml",
    connector: "DuckDB",
    connectorIcon: DuckDBIcon,
  },
  {
    name: "rill-openrtb-prog-ads",
    title: "OpenRTB Programmatic Ads",
    description: "Real-time Bidding (RTB) advertising",
    image: "/img/welcome-bg-openrtb.png",
    firstFile: "auction_explore.yaml",
    connector: "DuckDB",
    connectorIcon: DuckDBIcon,
  },
  {
    name: "rill-github-analytics",
    title: "Github Analytics",
    description: "A Git project's commit activity",
    image: "/img/welcome-bg-github-analytics.png",
    firstFile: "clickhouse_commits_explore.yaml",
    connector: "ClickHouse",
    connectorIcon: ClickHouseIcon,
  },
];
