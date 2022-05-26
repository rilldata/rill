import { line, area, curveLinear, curveStep } from "d3-shape";
import type { ScaleLinear } from "d3-scale";

export interface PlotConfig {
  width: number;
  height: number;
  devicePixelRatio: number;
  top: number;
  bottom: number;
  left: number;
  right: number;
  buffer: number;
  plotTop: number;
  plotBottom: number;
  plotLeft: number;
  plotRight: number;
  fontSize: number;
  textGap: number;
  id: string;
}

/**
 * Creates a string to be fed into the d attribute of a path,
 * producing a single path definition for one circle.
 * These completed, segmented arcs will not overlap in a way where
 * we can overplot if part of the same path.
 */
export function circlePath(cx: number, cy: number, r: number) {
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

function isDefined(yAccessor: string) {
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
  xScale: ScaleLinear<number, number>;
  yScale: ScaleLinear<number, number>;
  curve: string;
}

export function lineFactory(args: LineGeneratorArguments) {
  return (yAccessor: string) =>
    line()
      .x((d) => args.xScale(d[args.xAccessor]))
      .y((d) => Math.min(args.yScale.range()[0], args.yScale(d[yAccessor])))
      .curve(curves[args.curve] || curveLinear)
      .defined(isDefined(yAccessor));
}

export function areaFactory(args: LineGeneratorArguments) {
  return (yAccessor: string) =>
    area()
      .x((d) => ~~args.xScale(d[args.xAccessor]))
      .y0(~~args.yScale(0) + 0.5)
      .y1((d) => ~~args.yScale(d[yAccessor]))
      .curve(curves[args.curve] || curveLinear)
      .defined(isDefined(yAccessor));
}
