import { V1Operation } from "@rilldata/web-common/runtime-client";

export enum CriteriaOperations {
  GreaterThan = "GreaterThan",
  LessThan = "LessThan",
  // For backwards compatibility but not available as option
  GreaterThanOrEquals = "GreaterThanOrEquals",
  LessThanOrEquals = "LessThanOrEquals",
  IncreasesBy = "IncreasesBy",
  DecreasesBy = "DecreasesBy",
  ChangesBy = "ChangesBy",
  ShareOfTotalsGreaterThan = "ShareOfTotalsGreaterThan",
  ShareOfTotalsLessThan = "ShareOfTotalsLessThan",
}

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
    value: CriteriaOperations.GreaterThan,
    label: "is greater than",
  },
  {
    value: CriteriaOperations.LessThan,
    label: "is less than",
  },
  {
    value: CriteriaOperations.IncreasesBy,
    label: "increases by",
  },
  {
    value: CriteriaOperations.DecreasesBy,
    label: "decreases by",
  },
  {
    value: CriteriaOperations.ChangesBy,
    label: "changes by",
  },
  // TODO
  // {
  //   value: CriteriaOperations.ShareOfTotalsGreaterThan,
  //   label: "share of total is greater than",
  // },
  // {
  //   value: CriteriaOperations.ShareOfTotalsLessThan,
  //   label: "share of total is less than",
  // },
];

export const CriteriaGroupOptions = [
  { value: V1Operation.OPERATION_AND, label: "and" },
  { value: V1Operation.OPERATION_OR, label: "or" },
];
