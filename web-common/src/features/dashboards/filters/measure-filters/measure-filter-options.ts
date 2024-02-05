import { V1Operation } from "@rilldata/web-common/runtime-client";

export type MeasureFilterOption = {
  value: V1Operation;
  label: string;
  shortLabel: string;
};

export const MeasureFilterOptions: MeasureFilterOption[] = [
  {
    value: V1Operation.OPERATION_LT,
    label: "Less Than",
    shortLabel: "<",
  },
  {
    value: V1Operation.OPERATION_LTE,
    label: "Less Than Or Equals",
    shortLabel: "<=",
  },
  {
    value: V1Operation.OPERATION_GT,
    label: "Greater Than",
    shortLabel: ">",
  },
  {
    value: V1Operation.OPERATION_GTE,
    label: "Greater Than Or Equals",
    shortLabel: ">=",
  },
  {
    value: V1Operation.OPERATION_AND,
    label: "Between",
    shortLabel: "",
  },
  {
    value: V1Operation.OPERATION_OR,
    label: "Not Between",
    shortLabel: "",
  },
];
