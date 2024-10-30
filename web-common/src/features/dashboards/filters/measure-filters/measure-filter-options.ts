import { V1Operation } from "@rilldata/web-common/runtime-client";

export enum MeasureFilterType {
  Value = "Value",
  AbsoluteChange = "AbsoluteChange",
  PercentChange = "PercentChange",
  PercentOfTotal = "PercentOfTotal",
}

export enum MeasureFilterOperation {
  Equals = "OPERATION_EQ",
  NotEquals = "OPERATION_NEQ",
  GreaterThan = "OPERATION_GT",
  GreaterThanOrEquals = "OPERATION_GTE",
  LessThan = "OPERATION_LT",
  LessThanOrEquals = "OPERATION_LTE",
  Between = "Between",
  NotBetween = "NotBetween",
}

export const MeasureFilterToProtoOperation = {
  [MeasureFilterOperation.Equals]: V1Operation.OPERATION_EQ,
  [MeasureFilterOperation.NotEquals]: V1Operation.OPERATION_NEQ,
  [MeasureFilterOperation.GreaterThan]: V1Operation.OPERATION_GT,
  [MeasureFilterOperation.GreaterThanOrEquals]: V1Operation.OPERATION_GTE,
  [MeasureFilterOperation.LessThan]: V1Operation.OPERATION_LT,
  [MeasureFilterOperation.LessThanOrEquals]: V1Operation.OPERATION_LTE,
};
export const ProtoToMeasureFilterOperations: Partial<
  Record<V1Operation, MeasureFilterOperation>
> = {};
for (const MeasureFilterOperation in MeasureFilterToProtoOperation) {
  ProtoToMeasureFilterOperations[
    MeasureFilterToProtoOperation[MeasureFilterOperation]
  ] = MeasureFilterOperation;
}

export const MeasureFilterBaseTypeOptions = [
  {
    value: MeasureFilterType.Value,
    label: "value",
    shortLabel: "",
    description: "value",
  },
];
export const MeasureFilterComparisonTypeOptions = [
  {
    value: MeasureFilterType.PercentChange,
    label: "% change from",
    shortLabel: "% change",
    description: "% change",
  },
  {
    value: MeasureFilterType.AbsoluteChange,
    label: "change from",
    shortLabel: "change",
    description: "change",
  },
];
export const MeasureFilterPercentOfTotalOption = {
  value: MeasureFilterType.PercentOfTotal,
  label: "% of total",
  shortLabel: "% of total",
  description: "% of total",
};
export const AllMeasureFilterTypeOptions = [
  ...MeasureFilterBaseTypeOptions,
  ...MeasureFilterComparisonTypeOptions,
  MeasureFilterPercentOfTotalOption,
];

export const MeasureFilterOperationOptions = [
  {
    value: MeasureFilterOperation.GreaterThan,
    label: "Greater Than",
    shortLabel: ">",
  },
  {
    value: MeasureFilterOperation.GreaterThanOrEquals,
    label: "Greater Than Or Equals",
    shortLabel: ">=",
  },
  {
    value: MeasureFilterOperation.LessThan,
    label: "Less Than",
    shortLabel: "<",
  },
  {
    value: MeasureFilterOperation.LessThanOrEquals,
    label: "Less Than Or Equals",
    shortLabel: "<=",
  },
  {
    value: MeasureFilterOperation.Between,
    label: "Between",
    shortLabel: "",
  },
  {
    value: MeasureFilterOperation.NotBetween,
    label: "Not Between",
    shortLabel: "",
  },
];
// Full list with options not supported in filter pills just yet.
export const AllMeasureFilterOperationOptions = [
  ...MeasureFilterOperationOptions,
  {
    value: MeasureFilterOperation.Equals,
    label: "Equals",
    shortLabel: "=",
  },
  {
    value: MeasureFilterOperation.NotEquals,
    label: "Does Not Equals",
    shortLabel: "!=",
  },
];
