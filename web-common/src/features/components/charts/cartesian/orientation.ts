import type { VisualizationSpec } from "svelte-vega";
import type { ChartType } from "../types";
import type { CartesianChartSpec } from "./CartesianChartProvider";

export type ChartOrientation = "vertical" | "horizontal";

/** Chart types whose bars can be drawn horizontally. */
export const ORIENTATION_CHART_TYPES: ChartType[] = [
  "bar_chart",
  "stacked_bar",
  "stacked_bar_normalized",
];

export function supportsOrientation(type: ChartType): boolean {
  return ORIENTATION_CHART_TYPES.includes(type);
}

export function isHorizontal(
  config: Pick<CartesianChartSpec, "orientation">,
): boolean {
  return config.orientation === "horizontal";
}

type AnyRecord = Record<string, unknown>;

// Axis placement is expressed relative to the chart edge, so an axis that moves
// from the x channel to the y channel (or back) has to rotate its `orient`.
// The table is an involution, so it serves both directions.
const AXIS_ORIENT_TRANSPOSE: Record<string, string> = {
  bottom: "left",
  top: "right",
  left: "bottom",
  right: "top",
};

// Vega-Lite sort strings ("x", "-y", ...) name the channel to sort by, so they
// must follow the channels when those are swapped.
const SORT_CHANNEL_TRANSPOSE: Record<string, string> = {
  x: "y",
  "-x": "-y",
  y: "x",
  "-y": "-x",
};

/**
 * Transposes a cartesian bar spec built for the vertical orientation so that
 * the category runs along the y channel and the measure along the x channel.
 *
 * Builders always produce the vertical spec; horizontal charts are derived by
 * this single pass so that every builder (including comparison and
 * multi-measure modes) gets identical behavior. It deliberately touches only a
 * fixed set of properties:
 *   - `encoding.x` <-> `encoding.y` and `encoding.xOffset` <-> `encoding.yOffset`
 *   - `axis.orient` and string `sort` on the moved position defs
 *   - `axis.grid` so grid lines follow the measure axis, matching the vertical look
 *   - `mark.width` <-> `mark.height` (relative band sizes)
 *   - selection params' `select.encodings` ("x" <-> "y")
 * applied to the top-level encoding and recursively to every layer.
 * The top-level `width: "container"` is left alone on purpose.
 */
export function transposeCartesianSpec(
  spec: VisualizationSpec,
): VisualizationSpec {
  return transposeLayer(spec as AnyRecord) as VisualizationSpec;
}

function transposeLayer(layer: AnyRecord): AnyRecord {
  const out: AnyRecord = { ...layer };
  if (isRecord(layer.encoding)) {
    out.encoding = transposeEncoding(layer.encoding);
  }
  if (isRecord(layer.mark)) {
    out.mark = transposeMark(layer.mark);
  }
  if (Array.isArray(layer.params)) {
    out.params = (layer.params as unknown[]).map((p) =>
      isRecord(p) ? transposeParam(p) : p,
    );
  }
  if (Array.isArray(layer.layer)) {
    out.layer = (layer.layer as unknown[]).map((l) =>
      isRecord(l) ? transposeLayer(l) : l,
    );
  }
  return out;
}

function transposeEncoding(encoding: AnyRecord): AnyRecord {
  const { x, y, xOffset, yOffset, ...rest } = encoding;
  return {
    ...rest,
    ...(isRecord(y) && { x: transposePositionDef(y, "measure") }),
    ...(isRecord(x) && { y: transposePositionDef(x, "category") }),
    ...(yOffset !== undefined && { xOffset: yOffset }),
    ...(xOffset !== undefined && { yOffset: xOffset }),
  };
}

function transposePositionDef(
  def: AnyRecord,
  role: "category" | "measure",
): AnyRecord {
  const out: AnyRecord = { ...def };

  if (isRecord(def.axis)) {
    const axis: AnyRecord = { ...def.axis };
    if (typeof axis.orient === "string") {
      axis.orient = AXIS_ORIENT_TRANSPOSE[axis.orient] ?? axis.orient;
    }
    // The shared Vega config hides the grid on the x axis and shows it on the
    // y axis. Pin the grid to the measure axis so a horizontal chart keeps the
    // same look as a vertical one.
    if (axis.grid === undefined) {
      axis.grid = role === "measure";
    }
    out.axis = axis;
  }

  if (typeof def.sort === "string" && def.sort in SORT_CHANNEL_TRANSPOSE) {
    out.sort = SORT_CHANNEL_TRANSPOSE[def.sort];
  }

  return out;
}

function transposeMark(mark: AnyRecord): AnyRecord {
  const { width, height, ...rest } = mark;
  return {
    ...rest,
    ...(height !== undefined && { width: height }),
    ...(width !== undefined && { height: width }),
  };
}

function transposeParam(param: AnyRecord): AnyRecord {
  if (!isRecord(param.select) || !Array.isArray(param.select.encodings)) {
    return param;
  }
  return {
    ...param,
    select: {
      ...param.select,
      encodings: (param.select.encodings as unknown[]).map((channel) =>
        channel === "x" ? "y" : channel === "y" ? "x" : channel,
      ),
    },
  };
}

function isRecord(value: unknown): value is AnyRecord {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
