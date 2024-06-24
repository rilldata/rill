import { ChartField } from "./build-template";
import { singleLayerBaseSpec } from "./utils";

export function buildLine(
  timeField: ChartField,
  quantitativeField: ChartField,
  nominalField: ChartField | undefined,
) {
  const baseSpec = singleLayerBaseSpec();

  baseSpec.mark = { type: "line", clip: true };
  baseSpec.encoding = {
    x: { field: timeField.name, type: "temporal" },
    y: { field: quantitativeField.name, type: "quantitative" },
    ...(nominalField?.name && {
      color: { field: nominalField.name, type: "nominal" },
    }),
  };
  return baseSpec;
}
