import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type { ConnectError } from "@connectrpc/connect";
import type { Readable } from "svelte/store";
import type { GridItemHTMLElement, GridStackNode } from "gridstack";

export type CanvasSpecResponseStore = Readable<
  QueryObserverResult<CanvasResponse, ConnectError>
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
