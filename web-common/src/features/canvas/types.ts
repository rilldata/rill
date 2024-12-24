import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";

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
