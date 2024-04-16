import { singleLayerBaseSpec } from "./utils";

export function buildLine(
  timeField: string,
  quantitativeField: string,
  nominalField: string | undefined,
) {
  const baseSpec = singleLayerBaseSpec();

  baseSpec.mark = { type: "line", clip: true };
  baseSpec.encoding = {
    x: { field: timeField, type: "temporal" },
    y: { field: quantitativeField, type: "quantitative" },
    ...(nominalField && {
      color: { field: nominalField, type: "nominal" },
    }),
  };
  return baseSpec;
}
