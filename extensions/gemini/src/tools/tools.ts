import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { registerExportSheetTool, registerGenerateTool } from './generate.js';

/**
 * Registers all tools with the MCP server.
 */
export const registerTools = (server: McpServer) => {
  registerGenerateTool(server);
  registerExportSheetTool(server);
};
