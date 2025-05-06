import os
import sys
from datetime import datetime
from enum import Enum
from typing import Any, List, Optional

import httpx
from fastmcp import FastMCP
from pydantic import BaseModel

RILL_ADMIN_BASE_URL = os.getenv("RILL_ADMIN_BASE_URL") or "https://admin.rilldata.com"
RILL_ORGANIZATION_NAME = os.getenv("RILL_ORGANIZATION_NAME")
RILL_PROJECT_NAME = os.getenv("RILL_PROJECT_NAME")
RILL_SERVICE_TOKEN = os.getenv("RILL_SERVICE_TOKEN")

if not RILL_ORGANIZATION_NAME or not RILL_PROJECT_NAME or not RILL_SERVICE_TOKEN:
    print(
        "RILL_ORGANIZATION_NAME, RILL_PROJECT_NAME, and RILL_SERVICE_TOKEN must be set",
        file=sys.stderr,
    )
    raise SystemExit(1)

headers = {}
headers["Authorization"] = f"Bearer {RILL_SERVICE_TOKEN}"
admin_client = httpx.AsyncClient(
    base_url=RILL_ADMIN_BASE_URL,
    headers=headers,
)

runtime_client = httpx.AsyncClient()


async def get_project():
    response = await admin_client.get(
        f"/v1/organizations/{RILL_ORGANIZATION_NAME}/projects/{RILL_PROJECT_NAME}"
    )

    return response.json()


runtime_info = None


async def get_runtime_info(force_refresh=False):
    global runtime_info

    if runtime_info is None or force_refresh:
        project = await get_project()
        prod_deployment = project.get("prodDeployment", {})
        runtime_info = {
            "host": prod_deployment.get("runtimeHost"),
            "instance_id": prod_deployment.get("runtimeInstanceId"),
            "jwt": project.get("jwt"),
        }

    if runtime_info is None:
        raise ValueError("Failed to get runtime info")

    return runtime_info


def fix_dev_runtime_host(host: str) -> str:
    if host == "http://localhost:9091":
        return "http://localhost:8081"
    return host


rill_mcp = FastMCP(name="RillRuntimeServer")


@rill_mcp.tool()
async def list_metrics_views():
    """
    List all metrics views in the current project.
    """
    runtime_info = await get_runtime_info()
    host = fix_dev_runtime_host(runtime_info["host"])
    response = await runtime_client.get(
        f"{host}/v1/instances/{runtime_info['instance_id']}/resources?kind=rill.runtime.v1.MetricsView",
        headers={"Authorization": f"Bearer {runtime_info['jwt']}"},
    )

    names = [
        resource["meta"]["name"]["name"]
        for resource in response.json().get("resources", [])
    ]

    return names


class GetMetricsViewResourceRequest(BaseModel):
    name: str


@rill_mcp.tool()
async def get_metrics_view_spec(request: GetMetricsViewResourceRequest):
    """
    Retrieve the specification for a given metrics view, including available measures and dimensions.
    """
    runtime_info = await get_runtime_info()
    host = fix_dev_runtime_host(runtime_info["host"])
    response = await runtime_client.get(
        f"{host}/v1/instances/{runtime_info['instance_id']}/resource?name.name={request.name}&name.kind=rill.runtime.v1.MetricsView",
        headers={"Authorization": f"Bearer {runtime_info['jwt']}"},
    )

    response_json = response.json()

    try:
        valid_spec = response_json["resource"]["metricsView"]["state"]["validSpec"]
    except (KeyError, TypeError):
        valid_spec = {}

    return valid_spec


class GetMetricsViewTimeRangeSummaryRequest(BaseModel):
    metrics_view: str


@rill_mcp.tool()
async def get_metrics_view_time_range_summary(
    request: GetMetricsViewTimeRangeSummaryRequest,
):
    """
    Retrieve the total time range available for a given metrics view.

    Notes:
        All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
    """
    runtime_info = await get_runtime_info()
    host = fix_dev_runtime_host(runtime_info["host"])
    response = await runtime_client.post(
        f"{host}/v1/instances/{runtime_info['instance_id']}/queries/metrics-views/{request.metrics_view}/time-range-summary",
        headers={"Authorization": f"Bearer {runtime_info['jwt']}"},
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


class GetMetricsViewAggregationRequest(BaseModel):
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


@rill_mcp.tool()
async def get_metrics_view_aggregation(request: GetMetricsViewAggregationRequest):
    """
    Perform an arbitrary aggregation on a metrics view.

    Tip:
        Use the `sort` and `limit` parameters for best results and to avoid large, unbounded result sets.

    Examples:
        Get the total revenue by country and product category:

            {
                "metrics_view": "ecommerce_financials",
                "measures": [{"name": "total_revenue"}, {"name": "total_orders"}],
                "dimensions": [{"name": "country"}, {"name": "product_category"}],
                "time_range": {
                    "start": "2024-01-01T00:00:00Z",
                    "end": "2024-12-31T23:59:59Z"
                },
                "where": {
                    "op": "OPERATION_AND",
                    "exprs": [
                        {
                            "ident": "country",
                            "op": "OPERATION_IN",
                            "val": ["US", "CA", "GB"]
                        },
                        {
                            "ident": "product_category",
                            "op": "OPERATION_EQ",
                            "val": "Electronics"
                        }
                    ]
                },
                "sort": [{"name": "total_revenue", "desc": true}],
                "limit": "10"
            }
    """

    runtime_info = await get_runtime_info()
    host = fix_dev_runtime_host(runtime_info["host"])

    payload = request.model_dump(exclude={"metrics_view"}, mode="json")

    response = await runtime_client.post(
        f"{host}/v1/instances/{runtime_info['instance_id']}/queries/metrics-views/{request.metrics_view}/aggregation",
        headers={"Authorization": f"Bearer {runtime_info['jwt']}"},
        json=payload,
    )

    return response.json()


if __name__ == "__main__":
    rill_mcp.run()
