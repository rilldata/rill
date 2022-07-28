import { line, area, curveLinear, curveStep } from "d3-shape";
import type { ScaleLinear, ScaleTime } from "d3-scale";
import type { GraphicScale } from "./state/types";

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

interface LineGeneratorArguments {
  xAccessor: string;
  xScale: ScaleLinear<number, number> | ScaleTime<Date, number>;
  yScale: ScaleLinear<number, number> | ScaleTime<Date, number>;
  curve: string;
  pathDefined?: (datum: object, i: number, arr: ArrayLike<unknown>) => boolean;
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
      .curve(curves[args.curve] || curveLinear)
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
      .x((d) => ~~args.xScale(d[args.xAccessor]))
      .y0(~~args.yScale(0))
      .y1((d) => ~~args.yScale(d[yAccessor]))
      .curve(curves[args.curve] || curveLinear)
      .defined(args.pathDefined || pathDoesNotDropToZero(yAccessor));
}

/**
 * Return a list of ticks to be represented on the
 * axis or grid depending on axis-side, it's length and
 * the data type of field
 */
export function getTicks(
  xOrY: string,
  scale: GraphicScale,
  axisLength: number,
  isDate: boolean
) {
  const tickCount = ~~(axisLength / (xOrY === "x" ? 150 : 50));
  let ticks = scale.ticks(tickCount);

  if (ticks.length <= 1) {
    if (isDate) ticks = [...scale.domain(), ...ticks] as Date[];
    else ticks = scale.nice().domain();
  }

  return ticks;
}
