from datetime import datetime
from enum import Enum
from typing import Any, List, Optional, Union

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

    def dict(self, *args, **kwargs):
        d = super().dict(*args, **kwargs)
        # Convert datetime objects to ISO format strings
        d["start"] = self.start.isoformat()
        d["end"] = self.end.isoformat()
        return d


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
    where_sql: Optional[str] = None
    having: Optional[Expression] = None
    having_sql: Optional[str] = None
    limit: Optional[int] = None
    offset: Optional[int] = None
    exact: Optional[bool] = None
    fill_missing: Optional[bool] = None
    rows: Optional[bool] = False
