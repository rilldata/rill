import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1Operation } from "@rilldata/web-common/runtime-client";

export enum CompareWith {
  Value = "Value",
  Percent = "Percent",
}
export const CompareWithOptions = [
  {
    value: CompareWith.Value,
    label: "value",
  },
  {
    value: CompareWith.Percent,
    label: "percent",
  },
];

// TODO: should match measure filter. merge them once we add support for comparison based filters
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
    label: "does not equal",
    shortLabel: "!=",
  },
];
export const CriteriaOperationComparisonOptions = [
  ...CriteriaOperationOptions,
  {
    value: MeasureFilterOperation.IncreasesBy,
    label: "increases by",
  },
  {
    value: MeasureFilterOperation.DecreasesBy,
    label: "decreases by",
  },
  {
    value: MeasureFilterOperation.ChangesBy,
    label: "changes by",
  },
  // TODO
  // {
  //   value: MeasureFilterOperation.ShareOfTotalsGreaterThan,
  //   label: "share of total is greater than",
  // },
  // {
  //   value: MeasureFilterOperation.ShareOfTotalsLessThan,
  //   label: "share of total is less than",
  // },
];

export const CriteriaGroupOptions = [
  { value: V1Operation.OPERATION_AND, label: "and" },
  { value: V1Operation.OPERATION_OR, label: "or" },
];
