import { getContext, setContext } from "svelte";

/**
 * Lets chat surfaces start the browser-based OAuth flow for an external MCP server.
 * The flow runs on the admin service, which owns the user's session; only Rill Cloud provides the URL builder.
 */
export interface McpConnectContext {
  /** Returns the URL that starts the authorization flow for the connector. Opened in a popup. */
  connectUrl: (connector: string) => string;
}

const MCP_CONNECT_CONTEXT_KEY = Symbol("mcp-connect-context");

export function setMcpConnectContext(context: McpConnectContext): void {
  setContext(MCP_CONNECT_CONTEXT_KEY, context);
}

/**
 * Returns the MCP connect context, or undefined if no provider is an ancestor.
 * Surfaces without a provider (Rill Developer, embeds) can list servers but cannot start an authorization.
 */
export function getMcpConnectContext(): McpConnectContext | undefined {
  return getContext<McpConnectContext | undefined>(MCP_CONNECT_CONTEXT_KEY);
}
