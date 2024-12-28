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

export interface RowGroup {
  y: number;
  height: number;
  items: V1CanvasItem[];
}

export interface GridItem {
  position: [number, number]; // [x, y]
  size: [number, number]; // [width, height]
  node: any; // YAML node
}

export type DropPosition = "left" | "right" | "bottom" | "row";
