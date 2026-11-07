import { describe, expect, it } from "vitest";
import type { V1Message } from "@rilldata/web-common/runtime-client";
import { transformToBlocks } from "../core/messages/block-transform";
import {
  isMcpToolName,
  parseMcpAuthorizationRequired,
  parseMcpToolName,
} from "./mcp-servers";

describe("mcp-servers helpers", () => {
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

  it("parses the authorization required error envelope", () => {
    expect(
      parseMcpAuthorizationRequired(
        JSON.stringify({
          code: "mcp_authorization_required",
          connector: "linear",
          display_name: "Linear",
          message: "connect first",
        }),
      ),
    ).toEqual({ connector: "linear", displayName: "Linear" });
    expect(parseMcpAuthorizationRequired("not json")).toBeNull();
    expect(
      parseMcpAuthorizationRequired(JSON.stringify({ code: "other" })),
    ).toBeNull();
    expect(parseMcpAuthorizationRequired(undefined)).toBeNull();
  });
});

describe("block transform for external tools", () => {
  const call: V1Message = {
    id: "call-1",
    parentId: "root",
    role: "assistant",
    type: "call",
    tool: "mcp__linear__search_issues",
    contentType: "json",
    contentData: JSON.stringify({ query: "bugs" }),
  };

  it("renders an authorization required result as a connect card", () => {
    const result: V1Message = {
      id: "res-1",
      parentId: "call-1",
      role: "assistant",
      type: "result",
      tool: "mcp__linear__search_issues",
      contentType: "error",
      contentData: JSON.stringify({
        code: "mcp_authorization_required",
        connector: "linear",
        display_name: "Linear",
      }),
    };
    const blocks = transformToBlocks([call, result], false, false);
    expect(blocks).toHaveLength(1);
    expect(blocks[0]).toMatchObject({
      type: "mcp-auth-required",
      connector: "linear",
      displayName: "Linear",
    });
  });

  it("renders other external tool calls inline like built-in tools", () => {
    const result: V1Message = {
      id: "res-1",
      parentId: "call-1",
      role: "assistant",
      type: "result",
      tool: "mcp__linear__search_issues",
      contentType: "json",
      contentData: JSON.stringify({ content: "3 issues" }),
    };
    const blocks = transformToBlocks([call, result], false, false);
    expect(blocks).toHaveLength(1);
    expect(blocks[0].type).toBe("thinking");
  });
});
