/**
 * Icon mappings for AI tools.
 * Maps tool names to their corresponding Svelte icon components.
 */

import type { ComponentType } from "svelte";
import Chart from "../../../components/icons/Chart.svelte";
import List from "../../../components/icons/List.svelte";
import MetricsViewIcon from "../../../components/icons/MetricsViewIcon.svelte";
import Search from "../../../components/icons/Search.svelte";
import SummaryStatistics from "../../../components/icons/SummaryStatistics.svelte";

/**
 * Tool icon mapping. Tools not in this mapping will fall back to the Search icon.
 */
export const TOOL_ICONS: Record<string, ComponentType> = {
  list_metrics_views: List,
  get_metrics_view: MetricsViewIcon,
  query_metrics_view_summary: SummaryStatistics,
  query_metrics_view: Search,
  create_chart: Chart,
};

/**
 * Gets the icon component for a tool.
 * Falls back to Search icon if no mapping exists.
 */
export function getToolIcon(toolName: string | undefined): ComponentType {
  if (!toolName) return Search;
  return TOOL_ICONS[toolName] || Search;
}
