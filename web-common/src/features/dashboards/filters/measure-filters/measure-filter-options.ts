import { V1Operation } from "@rilldata/web-common/runtime-client";

export type MeasureFilterOption = {
  value: MeasureFilterOperation;
  label: string;
  shortLabel: string;
};

export enum MeasureFilterOperation {
  Equals = "OPERATION_EQ",
  NotEquals = "OPERATION_NEQ",
  GreaterThan = "OPERATION_GT",
  GreaterThanOrEquals = "OPERATION_GTE",
  LessThan = "OPERATION_LT",
  LessThanOrEquals = "OPERATION_LTE",
  Between = "Between",
  NotBetween = "NotBetween",
  IncreasesBy = "IncreasesBy",
  DecreasesBy = "DecreasesBy",
  ChangesBy = "ChangesBy",
  ShareOfTotalsGreaterThan = "ShareOfTotalsGreaterThan",
  ShareOfTotalsLessThan = "ShareOfTotalsLessThan",
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
export const ProtoToCompareMeasureFilterOperation = {
  [V1Operation.OPERATION_GT]: MeasureFilterOperation.IncreasesBy,
  [V1Operation.OPERATION_GTE]: MeasureFilterOperation.IncreasesBy,
  [V1Operation.OPERATION_LT]: MeasureFilterOperation.DecreasesBy,
  [V1Operation.OPERATION_LTE]: MeasureFilterOperation.DecreasesBy,
};

export const IsCompareMeasureFilterOperation = {
  [MeasureFilterOperation.IncreasesBy]: true,
  [MeasureFilterOperation.DecreasesBy]: true,
  [MeasureFilterOperation.ChangesBy]: true,
};

export const MeasureFilterOptions: MeasureFilterOption[] = [
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
export const AllMeasureFilterOptions = [
  ...MeasureFilterOptions,
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
