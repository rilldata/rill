import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { registerTools } from './tools/tools.js';

const server = new McpServer({
  name: 'Rill Report Generator',
  websiteUrl: 'https://rilldata.com',
  version: '0.1.0',
  title: 'Rill Report Generator',
});

// Register tools
registerTools(server);

// Starts the MCP server and connects it to the standard I/O transport.
async function start() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

// Start the server
void start();
