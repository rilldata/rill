import { V1Operation } from "@rilldata/web-common/runtime-client";

import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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

// Labels use lazy getters so they resolve in the active locale at access time
// (render) rather than freezing to the locale active when this module loaded.
// These option objects are only ever array-spread or read by property, so the
// getters survive aggregation in AllMeasureFilter*Options below.
export const MeasureFilterBaseTypeOptions = [
  {
    value: MeasureFilterType.Value,
    get label() {
      return m.filter_measure_type_value();
    },
    shortLabel: "",
    description: "value",
  },
];
export const MeasureFilterComparisonTypeOptions = [
  {
    value: MeasureFilterType.PercentChange,
    get label() {
      return m.filter_measure_type_percent_change_from();
    },
    shortLabel: "% change",
    description: "% change",
  },
  {
    value: MeasureFilterType.AbsoluteChange,
    get label() {
      return m.filter_measure_type_change_from();
    },
    shortLabel: "change",
    description: "change",
  },
];
export const MeasureFilterPercentOfTotalOption = {
  value: MeasureFilterType.PercentOfTotal,
  get label() {
    return m.filter_measure_type_percent_of_total();
  },
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
    get label() {
      return m.filter_measure_op_greater_than();
    },
    shortLabel: ">",
  },
  {
    value: MeasureFilterOperation.GreaterThanOrEquals,
    get label() {
      return m.filter_measure_op_greater_than_or_equals();
    },
    shortLabel: ">=",
  },
  {
    value: MeasureFilterOperation.LessThan,
    get label() {
      return m.filter_measure_op_less_than();
    },
    shortLabel: "<",
  },
  {
    value: MeasureFilterOperation.LessThanOrEquals,
    get label() {
      return m.filter_measure_op_less_than_or_equals();
    },
    shortLabel: "<=",
  },
  {
    value: MeasureFilterOperation.Between,
    get label() {
      return m.filter_measure_op_between();
    },
    shortLabel: "",
  },
  {
    value: MeasureFilterOperation.NotBetween,
    get label() {
      return m.filter_measure_op_not_between();
    },
    shortLabel: "",
  },
];
// Full list with options not supported in filter pills just yet.
export const AllMeasureFilterOperationOptions = [
  ...MeasureFilterOperationOptions,
  {
    value: MeasureFilterOperation.Equals,
    get label() {
      return m.filter_measure_op_equals();
    },
    shortLabel: "=",
  },
  {
    value: MeasureFilterOperation.NotEquals,
    get label() {
      return m.filter_measure_op_does_not_equal();
    },
    shortLabel: "!=",
  },
];
