import { extent, max, min } from "d3-array";
import type { ScaleLinear } from "d3-scale";
import { scaleLinear } from "d3-scale";
import { barplotPolyline } from "@rilldata/web-common/components/data-graphic/utils";
import { INTEGERS } from "@rilldata/web-common/lib/duckdb-data-types";
import type { NumericHistogramBinsBin } from "@rilldata/web-common/runtime-client";

interface PlotBounds {
  left: number;
  right: number;
  top: number;
  bottom: number;
}

interface SeparatorConfig {
  threshold?: number;
  size?: number;
}

export function createHistogramScales(
  data: NumericHistogramBinsBin[],
  type: string,
  plotBounds: PlotBounds,
  separatorConfig?: SeparatorConfig,
): {
  xScale: ScaleLinear<number, number>;
  yScale: ScaleLinear<number, number>;
  path: string;
} {
  const { left, right, top, bottom } = plotBounds;

  const xMin = min(data, (d) => d.low);
  const xMax = max(data, (d) => d.high);
  const [, yMax] = extent(data, (d) => d.count);

  const xScale = scaleLinear()
    .domain([xMin ?? 0, xMax ?? 1])
    .range([left, right]);
  const yScale = scaleLinear()
    .domain([0, yMax ?? 1])
    .range([bottom, top]);

  const threshold = separatorConfig?.threshold ?? 20;
  const size = separatorConfig?.size ?? 0.25;
  const separator = data?.length < threshold && INTEGERS.has(type) ? size : 0;

  const path = data
    ? barplotPolyline(data, xScale, yScale, separator, false, 1)
    : "";

  return { xScale, yScale, path };
}
