import { bisector } from "d3-array";
import type { ScaleLinear, ScaleTime } from "d3-scale";
import { area, curveLinear, curveStep, line } from "d3-shape";
import { timeFormat } from "d3-time-format";
import { curveStepExtended } from "./marks/curveStepExtended";

/**
 * Creates a string to be fed into the d attribute of a path,
 * producing a single path definition for one circle.
 * These completed, segmented arcs will not overlap in a way where
 * we can overplot if part of the same path.
 */
export function circlePath(cx: number, cy: number, r: number): string {
  return `
    M ${cx - r}, ${cy}
      a ${r},${r} 0 1,0 ${r * 2},0
      a ${r},${r} 0 1,0 ${-r * 2},0
    `;
}

const curves = {
  curveLinear,
  curveStep,
  curveStepExtended,
};

export function pathDoesNotDropToZero(yAccessor: string) {
  return (d, i: number, arr) => {
    return (
      !isNaN(d[yAccessor]) &&
      d[yAccessor] !== undefined &&
      // remove all zeroes where the previous or next value is also zero.
      // these do not add to our understanding.
      (!(i !== 0 && d[yAccessor] === 0 && arr[i - 1][yAccessor] === 0) ||
        !(
          i !== arr.length - 1 &&
          d[yAccessor] === 0 &&
          arr[i + 1][yAccessor] === 0
        ))
    );
  };
}

export interface PlotConfig {
  top: number;
  bottom: number;
  left: number;
  right: number;
  buffer: number;
  width: number;
  height: number;
  devicePixelRatio: number;
  plotTop: number;
  plotBottom: number;
  plotLeft: number;
  plotRight: number;
  fontSize: number;
  textGap: number;
  id: string;
}

interface LineGeneratorArguments {
  xAccessor: string;
  xScale:
    | ScaleLinear<number, number>
    | ScaleTime<Date, number>
    | ((d) => number);
  yScale:
    | ScaleLinear<number, number>
    | ScaleTime<Date, number>
    | ((d) => number);
  curve?: string;
  pathDefined?: (
    datum: object,
    i?: number,
    arr?: ArrayLike<unknown>,
  ) => boolean;
}

/**
 * A convenience function to generate a nice SVG path for a time series.
 * FIXME: rename to timeSeriesLineFactory.
 * FIXME: once we've gotten the data generics in place and threaded into components, let's make sure to type this.
 */
export function lineFactory(args: LineGeneratorArguments) {
  return (yAccessor: string) =>
    line()
      .x((d) => args.xScale(d[args.xAccessor]))
      .y((d) => args.yScale(d[yAccessor]))
      .curve(args.curve ? curves[args.curve] : curveLinear)
      .defined(args.pathDefined || pathDoesNotDropToZero(yAccessor));
}

/**
 * A convenience function to generate a nice SVG area path for a time series.
 * FIXME: rename to timeSeriesAreaFactory.
 * FIXME: once we've gotten the data generics in place and threaded into components, let's make sure to type this.
 */
export function areaFactory(args: LineGeneratorArguments) {
  return (yAccessor: string) =>
    area()
      .x((d) => args.xScale(d[args.xAccessor]))
      .y0(args.yScale(0))
      .y1((d) => args.yScale(d[yAccessor]))
      .curve(args.curve ? curves[args.curve] : curveLinear)
      .defined(args.pathDefined || pathDoesNotDropToZero(yAccessor));
}

/**
 * Generates an SVG path string for a histogram / bar plot.
 * Each bin is defined by a low/high x range and a y count value.
 * The path traces the outline of all non-zero bins, suitable for
 * both fill and stroke rendering.
 */
export function barplotPolyline(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: Record<string, any>[],
  xLow: string,
  xHigh: string,
  yAccessor: string,
  xScale: (v: number) => number,
  yScale: (v: number) => number,
  separator = 1,
  closeBottom = false,
  inflator = 1,
): string {
  if (!data?.length) return "";

  const baseline = yScale(0);

  const path = data.reduce((acc: string, datum, i) => {
    const count = datum[yAccessor] ?? 0;
    if (count === 0) return acc;

    const low = datum[xLow] ?? 0;
    const high = datum[xHigh] ?? 0;
    const x = xScale(low) + separator;
    const width = Math.max(0.5, xScale(high) - xScale(low) - separator * 2);
    const y = baseline * (1 - inflator) + yScale(count) * inflator;
    const barHeight = Math.min(
      baseline,
      baseline * inflator - yScale(count) * inflator,
    );
    const dropHeight = separator > 0 ? barHeight : 0;

    const prevIsZero = i > 0 && !data[i - 1][yAccessor];
    const nextIsZero = i < data.length - 1 && !data[i + 1][yAccessor];

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

  const lastNonZero = data.findLast((d) => d[yAccessor]);
  if (!lastNonZero) return "";

  const startX = xScale(data[0][xLow] ?? 0) + separator;
  const endX = xScale(lastNonZero[xHigh] ?? 0) - separator;
  return `M${startX},${baseline} ${path} ${endX},${baseline} `;
}

// This is function equivalent of WithBisector
export function bisectData<T>(
  value: Date,
  direction: "left" | "right" | "center",
  accessor: keyof T,
  data: ArrayLike<T>,
): { position: number; entry: T } {
  const bisect = bisector<T, unknown>((d) => d[accessor])[direction];
  const position = bisect(data, value);

  return {
    position,
    entry: data[position],
  };
}

/** For a scale domain returns a formatter for axis label and super label */
export function createTimeFormat(
  scaleDomain: [Date, Date],
  numberOfValues: number,
): [(d: Date) => string, ((d: Date) => string) | undefined] {
  const diff =
    Math.abs(scaleDomain[1]?.getTime() - scaleDomain[0]?.getTime()) / 1000;
  if (!diff) return [timeFormat("%d %b"), timeFormat("%Y")];
  const gap = diff / (numberOfValues - 1); // time gap between two consecutive values

  // If the gap is less than a second, format in milliseconds
  if (gap < 1) {
    return [timeFormat("%M:%S.%L"), timeFormat("%H %d %b %Y")];
  }
  // If the gap is less than a minute, format in seconds
  else if (gap < 60) {
    return [timeFormat("%M:%S"), timeFormat("%H %d %b %Y")];
  }
  // If the gap is less than 24 hours, format in hours and minutes
  else if (gap < 60 * 60 * 24) {
    return [timeFormat("%H:%M"), timeFormat("%d %b %Y")];
  }
  // If the gap is less than 30 days, format in days
  else if (gap < 60 * 60 * 24 * 30) {
    return [timeFormat("%b %d"), timeFormat("%Y")];
  }
  // If the gap is less than a year, format in months
  else if (gap < 60 * 60 * 24 * 365) {
    return [timeFormat("%b"), timeFormat("%Y")];
  }
  // Else format in years
  else {
    return [timeFormat("%Y"), undefined];
  }
}
