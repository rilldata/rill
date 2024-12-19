import type { GridItemHTMLElement, GridStackNode } from "gridstack";

export type Vector = [number, number];

export type GridstackDispatchEvents = {
  added: { event: Event; nodes: GridStackNode[] };
  change: { event: Event; nodes: GridStackNode[] };
  disable: { event: Event };
  drag: { event: Event; el: GridItemHTMLElement };
  dragstart: { event: Event; el: GridItemHTMLElement };
  dragstop: { event: Event; el: GridItemHTMLElement };
  dropped: {
    event: Event;
    previousNode: GridStackNode;
    newNode: GridStackNode;
  };
  enable: { event: Event };
  removed: { event: Event; el: GridItemHTMLElement };
  resize: { event: Event; el: GridItemHTMLElement };
  resizestart: { event: Event; el: GridItemHTMLElement };
  resizestop: { event: Event; el: GridItemHTMLElement };
  resizecontent: { event: Event; el: GridItemHTMLElement };
};

export type GridstackCallbackParams =
  | Event
  | GridStackNode
  | GridStackNode[]
  | GridItemHTMLElement;
