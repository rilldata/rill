import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1Operation } from "@rilldata/web-common/runtime-client";

// TODO: should match measure filter. merge them once we add support for comparison based filters
export const CriteriaOperationOptions = [
  {
    value: MeasureFilterOperation.GreaterThan,
    label: "> greater than",
    shortLabel: ">",
    description: "greater than",
  },
  {
    value: MeasureFilterOperation.GreaterThanOrEquals,
    label: ">= greater than or equals",
    shortLabel: ">=",
    description: "greater than or equal to",
  },
  {
    value: MeasureFilterOperation.LessThan,
    label: "< less than",
    shortLabel: "<",
    description: "less than",
  },
  {
    value: MeasureFilterOperation.LessThanOrEquals,
    label: "<= less than or equals",
    shortLabel: "<=",
    description: "less than or equal to",
  },
  {
    value: MeasureFilterOperation.Equals,
    label: "= equals",
    shortLabel: "=",
    description: "equal to",
  },
  {
    value: MeasureFilterOperation.NotEquals,
    label: "!= does not equal",
    shortLabel: "!=",
    description: "does not equal to",
  },
];

export const CriteriaGroupOptions = [
  { value: V1Operation.OPERATION_AND, label: "and" },
  { value: V1Operation.OPERATION_OR, label: "or" },
];
