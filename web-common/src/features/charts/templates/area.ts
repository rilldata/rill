import { ChartField } from "./build-template";
import { singleLayerBaseSpec } from "./utils";

export function buildArea(
  timeField: ChartField,
  quantitativeField: ChartField,
) {
  const baseSpec = singleLayerBaseSpec();

  baseSpec.mark = { type: "area", clip: true };
  baseSpec.encoding = {
    x: { field: timeField.name, type: "temporal" },
    y: { field: quantitativeField.name, type: "quantitative" },
  };

  return baseSpec;
}
