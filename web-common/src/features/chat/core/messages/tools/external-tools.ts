/**
 * Helpers for tools contributed by external MCP servers (connectors with driver "mcp").
 */

/** Prefix of tool names contributed by external MCP servers (matches runtime/ai.ExternalToolPrefix). */
export const MCP_TOOL_PREFIX = "mcp__";

/** Meta keys the runtime sets on external tools (see runtime/ai/external_tools.go). */
export const MCP_META_SERVER_DISPLAY_NAME = "rill/mcpServerDisplayName";
export const MCP_META_TOOL_NAME = "rill/mcpToolName";

export function isMcpToolName(toolName: string | undefined): boolean {
  return !!toolName && toolName.startsWith(MCP_TOOL_PREFIX);
}

/**
 * Splits a prefixed external tool name into its connector and tool parts.
 * Both parts are sanitized by the runtime, so prefer the tool's meta when it is available.
 */
export function parseMcpToolName(toolName: string): {
  connector: string;
  tool: string;
} {
  const rest = toolName.slice(MCP_TOOL_PREFIX.length);
  const idx = rest.indexOf("__");
  if (idx === -1) return { connector: rest, tool: "" };
  return { connector: rest.slice(0, idx), tool: rest.slice(idx + 2) };
}
