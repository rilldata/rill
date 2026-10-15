/**
 * Helpers for external MCP servers (connectors with driver "mcp") in the chat.
 *
 * The runtime lists the servers the current user may use and whether the user has authorized them.
 * Authorization itself happens in a popup served by the local server or the admin service (see mcp-connect-context.ts),
 * which posts a message back to the opener when it completes.
 */

import type { PartialMessage } from "@bufbuild/protobuf";
import type { MCPServer } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getRuntimeServiceDisconnectMCPServerMutationOptions,
  getRuntimeServiceListMCPServersQueryKey,
  getRuntimeServiceListMCPServersQueryOptions,
  getRuntimeServiceListToolsQueryKey,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { writable } from "svelte/store";

export type McpServer = PartialMessage<MCPServer>;

/** Prefix of tool names contributed by external MCP servers (matches runtime/ai.ExternalToolPrefix). */
export const MCP_TOOL_PREFIX = "mcp__";

/** Error code of the JSON error a tool call returns when the user must authorize the server (matches runtime/ai.ErrorCodeMCPAuthorizationRequired). */
export const MCP_AUTHORIZATION_REQUIRED_CODE = "mcp_authorization_required";

/** Meta keys the runtime sets on external tools (see runtime/ai/external_tools.go). */
export const MCP_META_SERVER = "rill/mcpServer";
export const MCP_META_SERVER_DISPLAY_NAME = "rill/mcpServerDisplayName";
export const MCP_META_TOOL_NAME = "rill/mcpToolName";

/**
 * Whether the external servers popover (the one place where servers are connected and disconnected) is open.
 * Other surfaces, such as the Connect card in a conversation, open it instead of running the authorization flow themselves.
 */
export const mcpServersPopoverOpen = writable(false);

export function openMcpServersPopover(): void {
  mcpServersPopoverOpen.set(true);
}

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

export interface McpAuthorizationRequired {
  connector: string;
  displayName: string;
}

/**
 * Parses the error content of a tool result. Returns the connector the user must authorize,
 * or null if the error is something else.
 */
export function parseMcpAuthorizationRequired(
  contentData: string | undefined,
): McpAuthorizationRequired | null {
  if (!contentData) return null;
  try {
    const parsed = JSON.parse(contentData) as Record<string, unknown> | null;
    if (
      !parsed ||
      parsed.code !== MCP_AUTHORIZATION_REQUIRED_CODE ||
      !parsed.connector
    ) {
      return null;
    }
    return {
      connector: String(parsed.connector),
      displayName: String(parsed.display_name || parsed.connector),
    };
  } catch {
    return null;
  }
}

/**
 * Query for the external MCP servers available to the current user.
 * Refetched on window focus because the authorization popup completes out of band.
 */
export function createMcpServersQuery(client: RuntimeClient) {
  return createQuery(
    getRuntimeServiceListMCPServersQueryOptions(
      client,
      {},
      {
        query: {
          staleTime: 30_000,
          refetchOnWindowFocus: true,
          select: (data) => data.servers ?? [],
        },
      },
    ),
  );
}

/** Invalidates the queries that change when a user connects or disconnects a server. */
export async function invalidateMcpServers(client: RuntimeClient) {
  await Promise.all([
    queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListMCPServersQueryKey(client.instanceId),
    }),
    queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListToolsQueryKey(client.instanceId),
    }),
  ]);
}

export function createDisconnectMcpServerMutation(client: RuntimeClient) {
  return createMutation(
    getRuntimeServiceDisconnectMCPServerMutationOptions(client, {
      onSuccess: () => invalidateMcpServers(client),
    }),
  );
}

export type McpOAuthResult = {
  status: "connected" | "error" | "closed";
  error?: string;
};

/**
 * Opens the authorization flow in a popup and resolves when it completes.
 * The callback page posts a `rill-mcp-oauth` message to the opener; if the user closes the popup
 * without finishing, the promise resolves with status "closed".
 */
export function openMcpConnectPopup(
  url: string,
  connector: string,
): Promise<McpOAuthResult> {
  return new Promise((resolve) => {
    const width = 600;
    const height = 720;
    const left = window.screenX + Math.max(0, (window.outerWidth - width) / 2);
    const top = window.screenY + Math.max(0, (window.outerHeight - height) / 2);
    const popup = window.open(
      url,
      "rill-mcp-oauth",
      `popup,width=${width},height=${height},left=${left},top=${top}`,
    );

    let settled = false;
    let closedPoll: ReturnType<typeof setInterval> | undefined = undefined;

    const finish = (result: McpOAuthResult) => {
      if (settled) return;
      settled = true;
      window.removeEventListener("message", onMessage);
      if (closedPoll) clearInterval(closedPoll);
      resolve(result);
    };

    // The callback page is served by the same origin that started the flow; ignore messages from anywhere else,
    // and from any window other than the popup, so another page cannot fake a successful connection.
    const expectedOrigin = new URL(url, window.location.href).origin;
    const onMessage = (event: MessageEvent) => {
      if (event.origin !== expectedOrigin) return;
      if (popup && event.source !== popup) return;
      const data = event.data as Record<string, unknown> | null;
      if (!data || data.type !== "rill-mcp-oauth") return;
      if (data.connector && data.connector !== connector) return;
      finish({
        status: data.status === "connected" ? "connected" : "error",
        error: data.error ? String(data.error) : undefined,
      });
    };
    window.addEventListener("message", onMessage);

    if (!popup) {
      // Popup blocked: fall back to navigating the current tab, the callback page links back.
      window.location.assign(url);
      finish({ status: "closed" });
      return;
    }
    closedPoll = setInterval(() => {
      if (popup.closed) {
        // Give a final message a moment to arrive before giving up.
        setTimeout(() => finish({ status: "closed" }), 300);
      }
    }, 500);
  });
}
