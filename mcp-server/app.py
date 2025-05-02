import asyncio
import logging
import os

from fastmcp import FastMCP
from modules.admin import admin_mcp
from modules.runtime import runtime_mcp

mcp = FastMCP(name="Rill MCP Server")


async def maybe_import_viz_server():
    openai_api_key = os.environ.get("OPENAI_API_KEY")
    if openai_api_key:
        from modules.viz import viz_mcp

        await mcp.import_server(prefix="viz", server=viz_mcp, tool_separator="_")
    else:
        logging.warning("OPENAI_API_KEY not set. Viz server will not be enabled.")


async def setup():
    await mcp.import_server(prefix="admin", server=admin_mcp, tool_separator="_")
    await mcp.import_server(prefix="runtime", server=runtime_mcp, tool_separator="_")
    await maybe_import_viz_server()


if __name__ == "__main__":
    asyncio.run(setup())
    mcp.run()
