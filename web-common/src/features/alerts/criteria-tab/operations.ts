import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1Operation } from "@rilldata/web-common/runtime-client";

// TODO: should match measure filter. merge them once we add support for comparison based filters
export const CriteriaOperationOptions = [
  {
    value: MeasureFilterOperation.GreaterThanOrEquals,
    label: ">=",
    shortLabel: ">=",
    description: "greater than or equal to",
  },
  {
    value: MeasureFilterOperation.GreaterThan,
    label: ">",
    shortLabel: ">",
    description: "greater than",
  },
  {
    value: MeasureFilterOperation.LessThanOrEquals,
    label: "<=",
    shortLabel: "<=",
    description: "less than or equal to",
  },
  {
    value: MeasureFilterOperation.LessThan,
    label: "<",
    shortLabel: "<",
    description: "less than",
  },
  {
    value: MeasureFilterOperation.Equals,
    label: "=",
    shortLabel: "=",
    description: "equal to",
  },
  {
    value: MeasureFilterOperation.NotEquals,
    label: "!=",
    shortLabel: "!=",
    description: "does not equal to",
  },
];

export const CriteriaGroupOptions = [
  { value: V1Operation.OPERATION_AND, label: "and" },
  { value: V1Operation.OPERATION_OR, label: "or" },
];
