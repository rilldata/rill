from datetime import datetime
from enum import Enum
from typing import Any, List, Optional

import httpx
from fastmcp import FastMCP
from pydantic import BaseModel

runtime_client = httpx.AsyncClient()


def fix_dev_runtime_host(host: str) -> str:
    if host == "http://localhost:9091":
        return "http://localhost:8081"
    return host


runtime_mcp = FastMCP(name="RillRuntimeServer")


class RuntimeRequest(BaseModel):
    host: str
    instance_id: str
    jwt: str


@runtime_mcp.tool()
async def list_metrics_views(request: RuntimeRequest):
    host = fix_dev_runtime_host(request.host)
    response = await runtime_client.get(
        f"{host}/v1/instances/{request.instance_id}/resources?kind=rill.runtime.v1.MetricsView",
        headers={"Authorization": f"Bearer {request.jwt}"},
    )

    # Extract the resource names from the response
    # This prevents us from returning too much data in the response
    names = [
        resource["meta"]["name"]["name"]
        for resource in response.json().get("resources", [])
    ]

    return names


class GetMetricsViewResourceRequest(RuntimeRequest):
    name: str


@runtime_mcp.tool()
async def get_metrics_view_spec(request: GetMetricsViewResourceRequest):
    host = fix_dev_runtime_host(request.host)
    response = await runtime_client.get(
        f"{host}/v1/instances/{request.instance_id}/resource?name.name={request.name}&name.kind=rill.runtime.v1.MetricsView",
        headers={"Authorization": f"Bearer {request.jwt}"},
    )

    response_json = response.json()
    try:
        valid_spec = response_json["metricsView"]["state"]["validSpec"]
    except (KeyError, TypeError):
        valid_spec = {}

    return valid_spec


class GetMetricsViewTimeRangeSummaryRequest(RuntimeRequest):
    metrics_view: str


@runtime_mcp.tool()
async def get_metrics_view_time_range_summary(
    request: GetMetricsViewTimeRangeSummaryRequest,
):
    host = fix_dev_runtime_host(request.host)
    response = await runtime_client.post(
        f"{host}/v1/instances/{request.instance_id}/queries/metrics-views/{request.metrics_view}/time-range-summary",
        headers={"Authorization": f"Bearer {request.jwt}"},
    )
    return response.json()


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


@runtime_mcp.tool()
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


if __name__ == "__main__":
    runtime_mcp.run()
