from typing import List

import httpx
from fastmcp import FastMCP

runtime_client = httpx.AsyncClient()


def fix_dev_runtime_host(host: str) -> str:
    if host == "http://localhost:9091":
        return "http://localhost:8081"
    return host


mcp = FastMCP(name="RillRuntimeServer")


@mcp.tool()
async def list_resources(host: str, instance_id: str, jwt: str):
    host = fix_dev_runtime_host(host)
    response = await runtime_client.get(
        f"{host}/v1/instances/{instance_id}/resources",
        headers={"Authorization": f"Bearer {jwt}"},
    )
    return response.json()


@mcp.tool()
async def get_metrics_view_time_range_summary(
    host: str,
    instance_id: str,
    jwt: str,
    metrics_view: str,
):
    host = fix_dev_runtime_host(host)
    response = await runtime_client.post(
        f"{host}/v1/instances/{instance_id}/queries/metrics-views/{metrics_view}/time-range-summary",
        headers={"Authorization": f"Bearer {jwt}"},
    )
    return response.json()


@mcp.tool()
async def get_metrics_view_aggregation(
    host: str,
    instance_id: str,
    jwt: str,
    metrics_view: str,
    dimensions: List[str],
    measures: List[str],
    # TODO: add other parameters
):
    host = fix_dev_runtime_host(host)

    response = await runtime_client.post(
        f"{host}/v1/instances/{instance_id}/queries/metrics-views/{metrics_view}/aggregation",
        headers={"Authorization": f"Bearer {jwt}"},
        json={
            "dimensions": [{"name": d} for d in dimensions],
            "measures": [{"name": m} for m in measures],
        },
    )
    return response.json()


mcp._mcp_server.instructions = "This server provides access to RillData Runtime APIs. "

if __name__ == "__main__":
    mcp.run()
