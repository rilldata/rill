/**
 * Icon mappings for AI tools.
 * Maps tool names to their corresponding Svelte icon components.
 */

import {
  FileSearch,
  FileText,
  FolderTree,
  Pencil,
  CornerDownRight,
} from "lucide-svelte";
import type { ComponentType } from "svelte";
import Chart from "../../../../../components/icons/Chart.svelte";
import List from "../../../../../components/icons/List.svelte";
import MetricsViewIcon from "../../../../../components/icons/MetricsViewIcon.svelte";
import Search from "../../../../../components/icons/Search.svelte";
import SummaryStatistics from "../../../../../components/icons/SummaryStatistics.svelte";
import { ToolName } from "../../types";

/**
 * Tool icon mapping. Tools not in this mapping will fall back to the Search icon.
 */
export const TOOL_ICONS: Record<string, ComponentType> = {
  // Analyst Agent tools
  [ToolName.LIST_METRICS_VIEWS]: List,
  [ToolName.GET_METRICS_VIEW]: MetricsViewIcon,
  [ToolName.QUERY_METRICS_VIEW_SUMMARY]: SummaryStatistics,
  [ToolName.QUERY_METRICS_VIEW]: Search,
  [ToolName.CREATE_CHART]: Chart,

  // Developer Agent tools
  [ToolName.SEARCH_FILES]: FileSearch,
  [ToolName.LIST_FILES]: FolderTree,
  [ToolName.READ_FILE]: FileText,
  [ToolName.WRITE_FILE]: Pencil,

  // Common tools
  [ToolName.NAVIGATE]: CornerDownRight,
};

/**
 * Gets the icon component for a tool.
 * Falls back to Search icon if no mapping exists.
 */
export function getToolIcon(toolName: string | undefined): ComponentType {
  if (!toolName) return Search;
  return TOOL_ICONS[toolName] || Search;
}
