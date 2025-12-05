/**
 * Tool Registry
 *
 * Central configuration for how tools render in the chat UI.
 *
 * Note: This registry is for tools that produce THINKING BLOCK content.
 * The `router_agent` is handled separately—it produces TEXT (the main
 * conversation), not tool call UI.
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { ToolName } from "../types";
import { createChartBlock, type ChartBlock } from "./chart/chart-block";

// =============================================================================
// RENDER MODES
// =============================================================================

/**
 * How a tool call renders in the thinking block UI:
 * - "inline": Shown as a collapsible tool call in thinking blocks
 * - "block": Shown in thinking, then produces a top-level block (chart, diff, etc.)
 * - "hidden": Not shown (internal orchestration agents)
 */
export type ToolRenderMode = "inline" | "block" | "hidden";

// =============================================================================
// TOOL CONFIGURATION
// =============================================================================

/**
 * Configuration for a tool's rendering behavior.
 */
export interface ToolConfig {
  renderMode: ToolRenderMode;
  /**
   * For block tools: factory function to create the block.
   * Receives the tool call message and its result message (if available).
   */
  createBlock?: (
    callMessage: V1Message,
    resultMessage: V1Message | undefined,
  ) => ChartBlock | null; // Union with other block types as they're added
}

/**
 * Default configuration for tools not explicitly registered.
 * Most tools render inline within thinking blocks.
 */
const DEFAULT_TOOL_CONFIG: ToolConfig = {
  renderMode: "inline",
};

/**
 * Tool configurations registry.
 *
 * To add a new block-level tool:
 * 1. Add the tool name to ToolName in types.ts
 * 2. Create the block type and factory in a new directory (e.g., file-diff/)
 * 3. Add an entry here with renderMode: "block" and the createBlock factory
 */
const TOOL_CONFIGS: Partial<Record<string, ToolConfig>> = {
  // Hidden tools - internal orchestration agents, not shown in UI
  // Note: router_agent is NOT here—it produces text, not thinking content
  [ToolName.ANALYST_AGENT]: { renderMode: "hidden" },
  [ToolName.DEVELOPER_AGENT]: { renderMode: "hidden" },

  // Block tools - shown in thinking, then render a top-level block
  [ToolName.CREATE_CHART]: {
    renderMode: "block",
    createBlock: createChartBlock,
  },

  // Future block tools:
  // [ToolName.WRITE_FILE]: {
  //   renderMode: "block",
  //   createBlock: createFileDiffBlock,
  // },

  // All other tools default to "inline" (shown in thinking blocks)
};

// =============================================================================
// PUBLIC API
// =============================================================================

/**
 * Get the configuration for a tool.
 */
export function getToolConfig(toolName: string | undefined): ToolConfig {
  if (!toolName) return DEFAULT_TOOL_CONFIG;
  return TOOL_CONFIGS[toolName] ?? DEFAULT_TOOL_CONFIG;
}

/**
 * Check if a tool renders as a top-level block.
 */
export function isBlockTool(toolName: string | undefined): boolean {
  return getToolConfig(toolName).renderMode === "block";
}

/**
 * Check if a tool is hidden from the UI.
 * Hidden tools are internal orchestration agents whose calls shouldn't be displayed.
 */
export function isHiddenTool(toolName: string | undefined): boolean {
  return getToolConfig(toolName).renderMode === "hidden";
}

/**
 * Check if a tool renders inline within thinking blocks.
 */
export function isInlineTool(toolName: string | undefined): boolean {
  return getToolConfig(toolName).renderMode === "inline";
}
