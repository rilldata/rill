import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1Operation } from "@rilldata/web-common/runtime-client";

// TODO: should match measure filter. merge them once we add support for comparison based filters
export const CriteriaOperationOptions = [
  {
    value: MeasureFilterOperation.GreaterThanOrEquals,
    label: ">=",
    shortLabel: ">=",
  },
  {
    value: MeasureFilterOperation.GreaterThan,
    label: ">",
    shortLabel: ">",
  },
  {
    value: MeasureFilterOperation.LessThanOrEquals,
    label: "<=",
    shortLabel: "<=",
  },
  {
    value: MeasureFilterOperation.LessThan,
    label: "<",
    shortLabel: "<",
  },
  {
    value: MeasureFilterOperation.Equals,
    label: "=",
    shortLabel: "=",
  },
  {
    value: MeasureFilterOperation.NotEquals,
    label: "!=",
    shortLabel: "!=",
  },
];

export const CriteriaGroupOptions = [
  { value: V1Operation.OPERATION_AND, label: "and" },
  { value: V1Operation.OPERATION_OR, label: "or" },
];
