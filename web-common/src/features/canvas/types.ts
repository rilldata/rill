import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";

export type Vector = [number, number];

export interface PositionedItem {
  x: number;
  y: number;
  width: number;
  height: number;
}

// Array of rows, where each row is an array of items
export type LayoutVector = V1CanvasItem[][];
