import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";

export type Vector = [number, number];

export interface PositionedItem {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface RowGroup {
  y: number;
  height: number;
  items: V1CanvasItem[];
}

export type DropPosition = "left" | "right" | "bottom" | "row";
