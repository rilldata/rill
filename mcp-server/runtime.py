from datetime import datetime
from enum import Enum
from typing import Any, List, Optional

import httpx
from fastmcp import FastMCP
from pydantic import BaseModel


class RuntimeRequest(BaseModel):
    host: str
    instance_id: str
    jwt: str


class GetMetricsViewTimeRangeSummaryRequest(RuntimeRequest):
    metrics_view: str


class MetricsViewAggregationDimension(BaseModel):
    name: str


class MetricsViewAggregationMeasure(BaseModel):
    name: str


class TimeRange(BaseModel):
    start: datetime
    end: datetime


class MetricsViewAggregationSort(BaseModel):
    name: str  # Dimension or measure name
    desc: Optional[bool] = None


class Operation(str, Enum):
    UNSPECIFIED = "OPERATION_UNSPECIFIED"
    EQ = "OPERATION_EQ"
    NEQ = "OPERATION_NEQ"
    LT = "OPERATION_LT"
    LTE = "OPERATION_LTE"
    GT = "OPERATION_GT"
    GTE = "OPERATION_GTE"
    OR = "OPERATION_OR"
    AND = "OPERATION_AND"
    IN = "OPERATION_IN"
    NIN = "OPERATION_NIN"
    LIKE = "OPERATION_LIKE"
    NLIKE = "OPERATION_NLIKE"


class Expression(BaseModel):
    ident: Optional[str] = None
    val: Optional[Any] = None
    cond: Optional["Condition"] = None
    subquery: Optional["Subquery"] = None


class Condition(BaseModel):
    op: Operation
    exprs: List[Expression]


class Subquery(BaseModel):
    dimension: Optional[str] = None
    measures: Optional[List[str]] = None
    where: Optional[Expression] = None
    having: Optional[Expression] = None


Expression.model_rebuild()  # This is needed for the forward references to work


class GetMetricsViewAggregationRequest(RuntimeRequest):
    metrics_view: str
    dimensions: List[MetricsViewAggregationDimension]
    measures: List[MetricsViewAggregationMeasure]
    sort: Optional[List[MetricsViewAggregationSort]] = None
    time_range: Optional[TimeRange] = None
    comparison_time_range: Optional[TimeRange] = None
    pivot_on: Optional[List[str]] = None
    where: Optional[Expression] = None
    # where_sql: Optional[str] = None
    having: Optional[Expression] = None
    # having_sql: Optional[str] = None
    limit: Optional[str] = None
    offset: Optional[str] = None
    exact: Optional[bool] = None
    fill_missing: Optional[bool] = None
    rows: Optional[bool] = False


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
async def list_metrics_views(request: RuntimeRequest):
    host = fix_dev_runtime_host(request.host)
    response = await runtime_client.get(
        f"{host}/v1/instances/{request.instance_id}/resources?kind=rill.runtime.v1.MetricsView",
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

    payload = request.model_dump(
        exclude={"host", "instance_id", "jwt", "metrics_view"}, mode="json"
    )

    response = await runtime_client.post(
        f"{host}/v1/instances/{request.instance_id}/queries/metrics-views/{request.metrics_view}/aggregation",
        headers={"Authorization": f"Bearer {request.jwt}"},
        json=payload,
    )

    return response.json()


mcp._mcp_server.instructions = """
## 🧠 Server Instructions: Rill Runtime MCP

This server exposes Rill Runtime APIs for querying **metrics views**—Rill's analytical units.

---

### Authentication

Every tool requires:
- `host`: Runtime server base URL  
- `instance_id`: Unique ID of the runtime  
- `jwt`: Bearer token for auth

Get these values from the `GetProject` tool in the **Rill Admin MCP**.

---

### How to Analyze a Metrics View

1. **Discover resources**  
   Use `list_metrics_views()` to list all metrics views in the project.

2. **Get time range**  
   Use `get_metrics_view_time_range_summary()` to fetch the full available time range for a metrics view.

3. **Run an aggregation query**  
   Use `get_metrics_view_aggregation()` to query a metrics view with:
   - `dimensions`: List of dimensions to group by
   - `measures`: List of measures to compute
   - Optional fields:
     - `sort`: Sorting rules
     - `time_range` / `comparison_time_range`: Filter by time
     - `where` / `having`: Use structured filters (`Expression`) or `*_sql` strings
     - `limit`, `offset`, `exact`, `fill_missing`, `rows`
  For best results, it's highly recommended to use the `sort` and `limit` fields to control the output.

"""


if __name__ == "__main__":
    mcp.run()
