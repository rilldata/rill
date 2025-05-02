import asyncio

from fastmcp import FastMCP
from modules.admin import admin_mcp
from modules.runtime import runtime_mcp
from modules.viz import viz_mcp

mcp = FastMCP(name="Rill MCP Server")


async def setup():
    await mcp.import_server(prefix="admin", server=admin_mcp, tool_separator="_")
    await mcp.import_server(prefix="runtime", server=runtime_mcp, tool_separator="_")
    await mcp.import_server(prefix="viz", server=viz_mcp, tool_separator="_")


if __name__ == "__main__":
    asyncio.run(setup())
    mcp.run()
