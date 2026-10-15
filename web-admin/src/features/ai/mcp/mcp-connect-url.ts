import { CANONICAL_ADMIN_URL } from "@rilldata/web-admin/client/http-client";

/**
 * Builds the URL that starts the browser-based authorization flow for an external MCP server.
 * The flow runs on the admin service, which owns the user's session and stores the resulting token.
 */
export function mcpConnectUrl(
  organization: string,
  project: string,
  connector: string,
  branch?: string,
): string {
  const params = new URLSearchParams({
    org: organization,
    project,
    connector,
    redirect: window.location.href,
  });
  if (branch) params.set("branch", branch);
  return `${CANONICAL_ADMIN_URL}/mcp/oauth/connect?${params.toString()}`;
}
