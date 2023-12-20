import { V1Operation } from "@rilldata/web-common/runtime-client";

export const MeasureFilterOptions = [
  {
    value: V1Operation.OPERATION_EQ,
    label: "=",
  },
  {
    value: V1Operation.OPERATION_NEQ,
    label: "!=",
  },
  {
    value: V1Operation.OPERATION_LT,
    label: "<",
  },
  {
    value: V1Operation.OPERATION_LTE,
    label: "<=",
  },
  {
    value: V1Operation.OPERATION_GT,
    label: ">",
  },
  {
    value: V1Operation.OPERATION_GTE,
    label: ">=",
  },
  {
    value: "b",
    label: "Between",
  },
  {
    value: "nb",
    label: "Not Between",
  },
];
