import type { V1Message } from "@rilldata/web-common/runtime-client";
import type { McpAuthorizationRequired } from "../../../connect/mcp-servers";

/**
 * Block shown when an external MCP tool call failed because the user has not authorized the server.
 * It renders a Connect card; once the user authorizes the server, the conversation continues automatically.
 */
export type McpAuthRequiredBlock = {
  type: "mcp-auth-required";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  connector: string;
  displayName: string;
};

export function createMcpAuthRequiredBlock(
  message: V1Message,
  resultMessage: V1Message,
  auth: McpAuthorizationRequired,
): McpAuthRequiredBlock {
  return {
    type: "mcp-auth-required",
    id: `mcp-auth-${message.id}`,
    message,
    resultMessage,
    connector: auth.connector,
    displayName: auth.displayName,
  };
}
