import type { Vector } from "./types";

export const vector = {
  add: (a: Vector, b: Vector): Vector => {
    return [a[0] + b[0], a[1] + b[1]];
  },
  subtract: (minuend: Vector, subtrahend: Vector): Vector => {
    return [minuend[0] - subtrahend[0], minuend[1] - subtrahend[1]];
  },
};
