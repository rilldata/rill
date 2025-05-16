from datetime import datetime
from enum import Enum
from typing import Any, List, Optional

from pydantic import BaseModel, model_validator


class GetMetricsViewResourceRequest(BaseModel):
    name: str


class GetMetricsViewTimeRangeSummaryRequest(BaseModel):
    metrics_view: str


class TimeGrain(str, Enum):
    UNSPECIFIED = "TIME_GRAIN_UNSPECIFIED"
    MILLISECOND = "TIME_GRAIN_MILLISECOND"
    SECOND = "TIME_GRAIN_SECOND"
    MINUTE = "TIME_GRAIN_MINUTE"
    HOUR = "TIME_GRAIN_HOUR"
    DAY = "TIME_GRAIN_DAY"
    WEEK = "TIME_GRAIN_WEEK"
    MONTH = "TIME_GRAIN_MONTH"
    QUARTER = "TIME_GRAIN_QUARTER"
    YEAR = "TIME_GRAIN_YEAR"


class MetricsViewAggregationDimension(BaseModel):
    name: str
    time_grain: Optional[TimeGrain] = None


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

    @model_validator(mode="after")
    def check_oneof(cls, values):
        fields = ["ident", "val", "cond", "subquery"]
        set_fields = [f for f in fields if getattr(values, f) is not None]
        if len(set_fields) > 1:
            raise ValueError(f"Only one of {fields} can be set, but got: {set_fields}")
        if len(set_fields) == 0:
            raise ValueError(f"One of {fields} must be set.")
        return values


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
