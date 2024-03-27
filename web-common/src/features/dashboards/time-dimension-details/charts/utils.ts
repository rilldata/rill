import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import type { TopLevelSpec } from "vega-lite";

export function sanitizeSpecForTDD(spec: TopLevelSpec): TopLevelSpec {
  if (!spec) return spec;

  const sanitizedSpec = { ...spec };

  if (sanitizedSpec.encoding) {
    // Remove titles from axes
    sanitizedSpec.encoding.x.axis = { title: "" };
    sanitizedSpec.encoding.y.axis = { title: "" };

    // Set timeUnit for x-axis
    sanitizedSpec.encoding.x.timeUnit = "hours";
  }

  return sanitizedSpec;
}
export function reduceDimensionData(dimensionData: DimensionDataItem[]) {
  return dimensionData
    .map((dimension) =>
      dimension.data.map((datum) => ({
        dimension: dimension.value,
        ...datum,
      })),
    )
    .flat();
}
