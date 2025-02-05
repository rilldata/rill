import type { QueryObserverResult } from "@rilldata/svelte-query";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";
import type { GridItemHTMLElement, GridStackNode } from "gridstack";

export type CanvasSpecResponseStore = Readable<
  QueryObserverResult<CanvasResponse, RpcStatus>
>;

export type GridstackDispatchEvents = {
  added: { event: Event; nodes: GridStackNode[] };
  change: { event: Event; nodes: GridStackNode[] };
  disable: { event: Event };
  dragstart: { event: Event; el: GridItemHTMLElement };
  drag: { event: Event; el: GridItemHTMLElement };
  dragstop: { event: Event; el: GridItemHTMLElement };
  dropped: {
    event: Event;
    previousNode: GridStackNode;
    newNode: GridStackNode;
  };
  enable: { event: Event };
  removed: { event: Event; el: GridItemHTMLElement };
  resizestart: { event: Event; el: GridItemHTMLElement };
  resize: { event: Event; el: GridItemHTMLElement };
  resizestop: { event: Event; el: GridItemHTMLElement };
};

export type GridstackCallbackParams =
  | Event
  | GridStackNode
  | GridStackNode[]
  | GridItemHTMLElement;
