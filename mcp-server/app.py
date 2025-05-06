import asyncio
import logging
import os

from fastmcp import FastMCP
from modules.rill import rill_mcp

mcp = FastMCP(name="Rill MCP Server")
mcp._mcp_server.instructions = """
## Rill MCP Server

This server exposes APIs for querying **metrics views** (Rill's analytical units).

### Workflow Overview
1. **List Metrics Views:** Use `list_metrics_views` to discover available metrics views in a project.
2. **Get Metrics View Spec:** Use `get_metrics_view_spec` to fetch a metrics view's spec. This is important to understand all the dimensions and measures in the metrics view.
3. **Get Time Range:** Use `get_metrics_view_time_range_summary` to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
4. **Query Aggregations:** Use `get_metrics_view_aggregation` to run queries.

In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
"""


async def maybe_import_viz_server():
    openai_api_key = os.environ.get("OPENAI_API_KEY")
    if openai_api_key:
        from modules.viz import viz_mcp

        await mcp.import_server(prefix="viz", server=viz_mcp, tool_separator="_")
    else:
        logging.warning(
            "OPENAI_API_KEY not set. The visualization server will not be enabled."
        )


async def setup():
    await mcp.import_server(prefix="rill", server=rill_mcp, tool_separator="_")
    await maybe_import_viz_server()


if __name__ == "__main__":
    asyncio.run(setup())
    mcp.run()
