import httpx
from fastmcp import FastMCP

from .types import (
    GetMetricsViewAggregationRequest,
    GetMetricsViewTimeRangeSummaryRequest,
    RuntimeRequest,
)

runtime_client = httpx.AsyncClient()


def fix_dev_runtime_host(host: str) -> str:
    if host == "http://localhost:9091":
        return "http://localhost:8081"
    return host


mcp = FastMCP(name="RillRuntimeServer")


@mcp.tool()
async def list_resources(request: RuntimeRequest):
    host = fix_dev_runtime_host(request.host)
    response = await runtime_client.get(
        f"{host}/v1/instances/{request.instance_id}/resources",
        headers={"Authorization": f"Bearer {request.jwt}"},
    )
    return response.json()


@mcp.tool()
async def get_metrics_view_time_range_summary(
    request: GetMetricsViewTimeRangeSummaryRequest,
):
    host = fix_dev_runtime_host(request.host)
    response = await runtime_client.post(
        f"{host}/v1/instances/{request.instance_id}/queries/metrics-views/{request.metrics_view}/time-range-summary",
        headers={"Authorization": f"Bearer {request.jwt}"},
    )
    return response.json()


@mcp.tool()
async def get_metrics_view_aggregation(request: GetMetricsViewAggregationRequest):
    host = fix_dev_runtime_host(request.host)

    # Convert the request to the expected API format
    dimensions = [{"name": d.name} for d in request.dimensions]
    measures = [{"name": m.name} for m in request.measures]

    payload = {
        "dimensions": dimensions,
        "measures": measures,
    }

    # Add optional parameters to payload if they exist
    if request.sort:
        payload["sort"] = [s.dict() for s in request.sort]
    if request.time_range:
        payload["timeRange"] = request.time_range.dict()
    if request.comparison_time_range:
        payload["comparisonTimeRange"] = request.comparison_time_range.dict()
    if request.pivot_on:
        payload["pivotOn"] = request.pivot_on
    if request.where:
        payload["where"] = request.where
    if request.where_sql:
        payload["whereSql"] = request.where_sql
    if request.having:
        payload["having"] = request.having
    if request.having_sql:
        payload["havingSql"] = request.having_sql
    if request.limit is not None:
        payload["limit"] = str(request.limit)
    if request.offset is not None:
        payload["offset"] = str(request.offset)
    if request.exact is not None:
        payload["exact"] = request.exact
    if request.fill_missing is not None:
        payload["fillMissing"] = request.fill_missing
    if request.rows is not None:
        payload["rows"] = request.rows

    response = await runtime_client.post(
        f"{host}/v1/instances/{request.instance_id}/queries/metrics-views/{request.metrics_view}/aggregation",
        headers={"Authorization": f"Bearer {request.jwt}"},
        json=payload,
    )

    return response.json()


mcp._mcp_server.instructions = "This server provides access to RillData Runtime APIs. "

if __name__ == "__main__":
    mcp.run()
