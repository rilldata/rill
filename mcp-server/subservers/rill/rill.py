from fastmcp import FastMCP
from subservers.rill.admin_client import admin_client
from subservers.rill.pydantic_models import (
    GetMetricsViewAggregationRequest,
    GetMetricsViewResourceRequest,
    GetMetricsViewTimeRangeSummaryRequest,
)
from subservers.rill.utils import prune

rill_mcp = FastMCP(name="RillRuntimeServer")


@rill_mcp.tool()
async def list_metrics_views():
    """
    List all metrics views in the current project.
    """
    response = await admin_client.get(
        f"/runtime/resources?kind=rill.runtime.v1.MetricsView"
    )

    names = [
        resource["meta"]["name"]["name"]
        for resource in response.json().get("resources", [])
    ]

    return names


@rill_mcp.tool()
async def get_metrics_view_spec(request: GetMetricsViewResourceRequest):
    """
    Retrieve the specification for a given metrics view, including available measures and dimensions.
    """
    response = await admin_client.get(
        f"/runtime/resource?name.kind=rill.runtime.v1.MetricsView&name.name={request.name}",
    )

    response_json = response.json()

    try:
        valid_spec = response_json["resource"]["metricsView"]["state"]["validSpec"]
    except (KeyError, TypeError):
        valid_spec = {}

    return prune(valid_spec)


@rill_mcp.tool()
async def get_metrics_view_time_range_summary(
    request: GetMetricsViewTimeRangeSummaryRequest,
):
    """
    Retrieve the total time range available for a given metrics view.

    Notes:
        All subsequent queries of the metrics view should be constrained to this time range to ensure accurate results.
    """
    response = await admin_client.post(
        f"/runtime/queries/metrics-views/{request.metrics_view}/time-range-summary",
    )
    return response.json()


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
                    "cond": {
                        "op": "OPERATION_AND",
                        "exprs": [
                            {
                                "cond": {
                                    "op": "OPERATION_IN",
                                    "exprs": [
                                        {"ident": "country"},
                                        {"val": ["US", "CA", "GB"]}
                                    ]
                                }
                            },
                            {
                                "cond": {
                                    "op": "OPERATION_EQ",
                                    "exprs": [
                                        {"ident": "product_category"},
                                        {"val": "Electronics"}
                                    ]
                                }
                            }
                        ]
                    },
                },
                "sort": [{"name": "total_revenue", "desc": true}],
                "limit": "10"
            }

        Get the total revenue by country, grouped by month:

            {
                "metrics_view": "ecommerce_financials",
                "measures": [{"name": "total_revenue"}],
                "dimensions": [
                    {"name": "transaction_timestamp", "time_grain": "TIME_GRAIN_MONTH"}
                    {"name": "country"},
                ],
                "time_range": {
                    "start": "2024-01-01T00:00:00Z",
                    "end": "2024-12-31T23:59:59Z"
                },
                "sort": [
                    {"name": "transaction_timestamp"},
                    {"name": "total_revenue", "desc": true},
                ],
            }
    """

    payload = request.model_dump(
        exclude={"metrics_view"}, exclude_none=True, mode="json"
    )

    response = await admin_client.post(
        f"/runtime/queries/metrics-views/{request.metrics_view}/aggregation",
        json=payload,
    )

    return response.json()


if __name__ == "__main__":
    rill_mcp.run()
