import type { Vector } from "./types";

export const vector = {
  add: (add: Vector, initial: Vector): Vector => {
    return [add[0] + initial[0], add[1] + initial[1]];
  },
  multiply: (vector: Vector, multiplier: Vector): Vector => {
    // console.log("multiply");
    return [vector[0] * multiplier[0], vector[1] * multiplier[1]];
  },
  subtract: (minuend: Vector, subtrahend: Vector): Vector => {
    return [minuend[0] - subtrahend[0], minuend[1] - subtrahend[1]];
  },
  absolute: (vector: Vector): Vector => {
    return [Math.abs(vector[0]), Math.abs(vector[1])];
  },
  divide: (vector: Vector, divisor: Vector): Vector => {
    return [vector[0] / divisor[0], vector[1] / divisor[1]];
  },
};
