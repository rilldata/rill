import { Plug } from "lucide-svelte";
import { describe, expect, it } from "vitest";
import { isMcpToolName, parseMcpToolName } from "./external-tools";
import { getToolDisplayName } from "./tool-display-names";
import { getToolIcon } from "./tool-icons";

describe("external tools", () => {
  it("recognizes and parses external tool names", () => {
    expect(isMcpToolName("mcp__linear__search_issues")).toBe(true);
    expect(isMcpToolName("query_metrics_view")).toBe(false);
    expect(isMcpToolName(undefined)).toBe(false);
    expect(parseMcpToolName("mcp__linear__search_issues")).toEqual({
      connector: "linear",
      tool: "search_issues",
    });
    expect(parseMcpToolName("mcp__linear__get__issue")).toEqual({
      connector: "linear",
      tool: "get__issue",
    });
  });

  it("labels external tools with the server and tool name", () => {
    const tools = [
      {
        name: "mcp__linear__search_issues",
        meta: {
          "rill/mcpServerDisplayName": "Linear",
          "rill/mcpToolName": "search_issues",
        },
      },
    ];
    expect(getToolDisplayName("mcp__linear__search_issues", true, tools)).toBe(
      "Linear: search_issues",
    );
    // Without the tool's meta, the label is parsed from the name
    expect(getToolDisplayName("mcp__linear__get_issue", false, tools)).toBe(
      "linear: get_issue",
    );
  });

  it("uses the plug icon for external tools", () => {
    expect(getToolIcon("mcp__linear__search_issues")).toBe(Plug);
  });
});
