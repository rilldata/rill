# Rill MCP Go Server

This directory contains the Go implementation of the Rill Metrics Control Plane (MCP) server using the [`mcp-go`](https://github.com/mark3labs/mcp-go) library. This server replaces the previous Python-based MCP server and is designed to expose Rill's metrics and visualization tools to LLM agents and desktop copilots such as Claude Desktop.

## Features
- Implements the MCP protocol for tool and resource discovery
- Exposes Rill and visualization tools to LLM agents
- Easily extensible with new tools and handlers

## Directory Structure
- `server.go`: Main MCP server implementation and handler registration
- `tools.go`: Tool handler implementations for Rill and visualization
- `cmd/main.go`: Entrypoint to run the MCP server

## Building and Running

1. **Build the server:**
   ```sh
   cd runtime/server/mcp/cmd
   go build
   ```
   This will produce a binary named `cmd` (or `cmd.exe` on Windows).

2. **Run the server:**
   ```sh
   ./cmd
   ```
   The server will start and listen for MCP protocol requests on stdio (suitable for desktop copilots).

## Configuring Claude Desktop to Use the Rill MCP Server

1. **Start the Rill MCP server** (see above).
2. **Open Claude Desktop.**
3. Go to **Settings** > **Advanced** > **Custom Tool Servers** (or similar, depending on Claude Desktop version).
4. **Add a new custom tool server**:
   - **Type:** Local MCP server (or "Custom MCP server")
   - **Command:** Path to the built MCP server binary (e.g., `/Users/nishant/dev/rill/rill/runtime/server/mcp/cmd/cmd`)
   - **Arguments:** (leave blank unless you have custom flags)
   - **Working Directory:** `/Users/nishant/dev/rill/rill/runtime/server/mcp/cmd` (or wherever you built the binary)
   - **Environment Variables:** (optional, set any needed for Rill)
5. **Save** and **enable** the custom tool server.
6. Claude Desktop should now be able to discover and use Rill's tools via the MCP protocol.

## Extending the Server
- Add new tools by implementing handlers in `tools.go` and registering them in `server.go`.
- See the [`mcp-go` documentation](https://github.com/mark3labs/mcp-go) for protocol and handler details.

## Development Notes
- The server is designed for integration with LLM agents and desktop copilots, not for direct HTTP/gRPC use.
- For questions or contributions, open an issue or PR in the main Rill repository. 