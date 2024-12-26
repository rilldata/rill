import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";
import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";

export type Vector = [number, number];

export type CanvasSpecResponseStore = Readable<
  QueryObserverResult<CanvasResponse, RpcStatus>
>;

export interface PositionedItem {
  x: number;
  y: number;
  width: number;
  height: number;
}

// Array of rows, where each row is an array of items
export type LayoutVector = V1CanvasItem[][];
