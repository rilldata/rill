import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1Operation } from "@rilldata/web-common/runtime-client";

// TODO: should match measure filter. remove this once that is merged to main
export const CriteriaOperationOptions = [
  {
    value: MeasureFilterOperation.GreaterThanOrEquals,
    label: "greater than or equals",
    shortLabel: ">=",
  },
  {
    value: MeasureFilterOperation.LessThan,
    label: "less than",
    shortLabel: "<",
  },
  {
    value: MeasureFilterOperation.Equals,
    label: "equals",
    shortLabel: "=",
  },
  {
    value: MeasureFilterOperation.NotEquals,
    label: "does not equals",
    shortLabel: "!=",
  },
];

export const CriteriaGroupOptions = [
  { value: V1Operation.OPERATION_AND, label: "and" },
  { value: V1Operation.OPERATION_OR, label: "or" },
];
