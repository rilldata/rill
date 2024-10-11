import type { PivotPos } from "./types";

export function range(x0: number, x1: number, f: (x: number) => any) {
  return Array.from(Array(x1 - x0).keys()).map((x) => f(x + x0));
}

export const isEmptyPos = (pos: PivotPos) =>
  pos.x0 === 0 && pos.x1 === 0 && pos.y0 === 0 && pos.y1 === 0;
