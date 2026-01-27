/**
 * Tool Registry
 *
 * Central configuration for how tool calls render in the chat UI.
 *
 * Note: `router_agent` is NOT in this registry—its output goes to TextBlocks
 * (the main conversation), not tool call UI.
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { ToolName } from "../../types";
import { createChartBlock, type ChartBlock } from "../chart/chart-block";
import {
  createFileDiffBlock,
  type FileDiffBlock,
} from "../file-diff/file-diff-block";
import { goto } from "$app/navigation";
import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
import {
  createGenericBlock,
  type GenericBlock,
} from "@rilldata/web-common/features/chat/core/messages/generic/generic-block.ts";

// =============================================================================
// RENDER MODES
// =============================================================================

/**
 * How a tool call renders in the UI:
 * - "inline": Shown as a collapsible tool call in thinking blocks
 * - "block": Renders as a standalone block with its own header (chart, diff, etc.)
 * - "hidden": Not shown (internal orchestration agents)
 */
export type ToolRenderMode = "inline" | "block" | "hidden";

// =============================================================================
// TOOL CONFIGURATION
// =============================================================================

/** Block types that can be created by tools */
export type ToolBlockType = ChartBlock | FileDiffBlock | GenericBlock;

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
  ) => ToolBlockType | null;

  /** Used to process any UI action or side effects from tool calls. */
  onResult?: (callMessage: V1Message) => void;
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
  [ToolName.DEVELOP_MODEL]: { renderMode: "hidden" },
  [ToolName.DEVELOP_METRICS_VIEW]: { renderMode: "hidden" },

  // Block tools - render as standalone blocks with their own header
  [ToolName.CREATE_CHART]: {
    renderMode: "block",
    createBlock: createChartBlock,
  },
  [ToolName.WRITE_FILE]: {
    renderMode: "block",
    createBlock: createFileDiffBlock,
  },

  [ToolName.NAVIGATE]: {
    renderMode: "block",
    createBlock: createGenericBlock,
    onResult: handleNavigateToolCall,
  },

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
 * Check if a tool is hidden from the UI.
 * Hidden tools are internal orchestration agents whose calls shouldn't be displayed.
 */
export function isHiddenTool(toolName: string | undefined): boolean {
  return getToolConfig(toolName).renderMode === "hidden";
}

function handleNavigateToolCall(callMessage: V1Message) {
  if (!callMessage.contentData) return;
  try {
    const content = JSON.parse(callMessage.contentData);
    if (!content.kind || !content.name) return;
    switch (content.kind) {
      case "file":
        void goto(`/files${addLeadingSlash(content.name)}`);
    }
  } catch (err) {
    console.error(err);
  }
}
