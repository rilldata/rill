import { line, area, curveLinear, curveStep } from "d3-shape";
import type { scaleLinear } from "d3-scale";

const curves = {
  curveLinear,
  curveStep,
};

function isDefined(yAccessor: string) {
  return (d: any, i: number, arr: any[]) => {
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
  xScale: scaleLinear;
  yScale: scaleLinear;
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
