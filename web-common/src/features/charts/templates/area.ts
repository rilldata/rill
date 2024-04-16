import { singleLayerBaseSpec } from "./utils";

export function buildArea(timeField: string, quantitativeField: string) {
  const baseSpec = singleLayerBaseSpec();

  baseSpec.mark = { type: "area", clip: true };
  baseSpec.encoding = {
    x: { field: timeField, type: "temporal" },
    y: { field: quantitativeField, type: "quantitative" },
  };

  return baseSpec;
}
