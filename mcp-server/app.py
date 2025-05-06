import asyncio
import logging
import os

from fastmcp import FastMCP
from modules.admin import admin_mcp
from modules.runtime import runtime_mcp

mcp = FastMCP(name="Rill MCP Server")
mcp._mcp_server.instructions = """
## Rill MCP Server - API Usage Guide

This server exposes APIs for querying **metrics views** (Rill's analytical units).

### Workflow Overview
1. **List Projects:** Use `list_projects` to get all project names.
2. **Get Runtime Details:** Use `get_project_runtime` to retrieve `host`, `instance_id`, and `jwt` for a project. These are required for all subsequent tool calls.
3. **List Metrics Views:** Use `list_metrics_views` to discover available metrics views in a project.
4. **Get Metrics View Spec:** Use `get_metrics_view_spec` to fetch a metrics view's spec. This is important to understand all the dimensions and measures in the metrics view.
5. **Get Time Range:** Use `get_metrics_view_time_range_summary` to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
6. **Query Aggregations:** Use `get_metrics_view_aggregation` to run queries with:
   - `dimensions`: Grouping fields
   - `measures`: Metrics to compute
   - Optional: `sort`, `limit`, `offset`, `time_range`, `comparison_time_range`, `where`, `having`, `exact`, `fill_missing`, `rows`
   - **Tip:** Use `sort` and `limit` for best results.

### Authentication
- All runtime tools require: `host`, `instance_id`, and `jwt` (from `get_project_runtime`).
- Always fetch these from the admin server before making runtime requests.
"""


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
