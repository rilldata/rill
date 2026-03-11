import type { NumericHistogramBinsBin } from "@rilldata/web-common/runtime-client";
import { area, curveLinear, line } from "d3-shape";
import type { Area, CurveFactory, Line } from "d3-shape";

/**
 * Creates a configured d3 line generator.
 * Defaults to curveLinear; pass `defined` to skip null/invalid points.
 */
export function createLineGenerator<T>(args: {
  x: (d: T, i: number, data: T[]) => number;
  y: (d: T, i: number, data: T[]) => number;
  defined?: (d: T, i: number, data: T[]) => boolean;
  curve?: CurveFactory;
}): Line<T> {
  const gen = line<T>()
    .x(args.x)
    .y(args.y)
    .curve(args.curve ?? curveLinear);
  if (args.defined) gen.defined(args.defined);
  return gen;
}

/**
 * Creates a configured d3 area generator.
 * y0/y1 each accept a constant number or an accessor function.
 * Defaults to curveLinear; pass `defined` to skip null/invalid points.
 */
export function createAreaGenerator<T>(args: {
  x: (d: T, i: number, data: T[]) => number;
  y0: number | ((d: T, i: number, data: T[]) => number);
  y1: number | ((d: T, i: number, data: T[]) => number);
  defined?: (d: T, i: number, data: T[]) => boolean;
  curve?: CurveFactory;
}): Area<T> {
  const gen = area<T>()
    .x(args.x)
    .curve(args.curve ?? curveLinear);
  // The typeof narrowing is required: d3's .y0()/.y1() have separate overloads
  // for number vs function, and TypeScript won't accept the union directly.
  if (typeof args.y0 === "number") gen.y0(args.y0);
  else gen.y0(args.y0);
  if (typeof args.y1 === "number") gen.y1(args.y1);
  else gen.y1(args.y1);
  if (args.defined) gen.defined(args.defined);
  return gen;
}

/**
 * Filter predicate that removes consecutive zero values from a path.
 * Zeroes are kept if at least one neighbor is non-zero.
 */
export function pathDoesNotDropToZero<T>(yAccessor: keyof T) {
  return (d: T, i: number, arr: T[]): boolean => {
    return (
      (typeof d[yAccessor] !== "number" || !isNaN(d[yAccessor])) &&
      d[yAccessor] !== undefined &&
      (!(i !== 0 && d[yAccessor] === 0 && arr[i - 1][yAccessor] === 0) ||
        !(
          i !== arr.length - 1 &&
          d[yAccessor] === 0 &&
          arr[i + 1][yAccessor] === 0
        ))
    );
  };
}

/**
 * Generates an SVG path string for a histogram / bar plot.
 * Each bin is defined by a low/high x range and a y count value.
 * The path traces the outline of all non-zero bins, suitable for
 * both fill and stroke rendering.
 */
export function barplotPolyline(
  data: NumericHistogramBinsBin[],
  xScale: (v: number) => number,
  yScale: (v: number) => number,
  separator = 1,
  closeBottom = false,
  inflator = 1,
): string {
  if (!data?.length) return "";

  const baseline = yScale(0);

  const path = data.reduce((acc: string, datum, i) => {
    const count = datum.count ?? 0;
    if (count === 0) return acc;

    const low = datum.low ?? 0;
    const high = datum.high ?? 0;
    const x = xScale(low) + separator;
    const width = Math.max(0.5, xScale(high) - xScale(low) - separator * 2);
    const y = baseline * (1 - inflator) + yScale(count) * inflator;
    const barHeight = Math.min(
      baseline,
      baseline * inflator - yScale(count) * inflator,
    );
    const dropHeight = separator > 0 ? barHeight : 0;

    const prevIsZero = i > 0 && !data[i - 1].count;
    const nextIsZero = i < data.length - 1 && !data[i + 1].count;

    // Move to the bottom-left of this bar
    let move: string;
    if (separator === 0 && prevIsZero) {
      move = `M${x},${y + barHeight}`;
    } else if (separator > 0 || i === 0) {
      move = `${i !== 0 ? "M" : ""}${x},${y + dropHeight}`;
    } else {
      move = "";
    }

    const topLeft = `${x},${y}`;
    const topRight = `${x + width},${y}`;
    const bottomRight =
      separator > 0 || nextIsZero
        ? `${x + width},${y + (separator > 0 ? dropHeight : barHeight)}`
        : "";
    const close = closeBottom ? `${x},${y + dropHeight}` : "";

    return acc + `${move} ${topLeft} ${topRight} ${bottomRight} ${close} `;
  }, " ");

  const lastNonZero = data.findLast((d) => d.count);
  if (!lastNonZero) return "";

  const startX = xScale(data[0].low ?? 0) + separator;
  const endX = xScale(lastNonZero.high ?? 0) - separator;
  return `M${startX},${baseline} ${path} ${endX},${baseline} `;
}
