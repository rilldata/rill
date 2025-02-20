import type {
  Alignment,
  Location,
} from "@rilldata/web-common/lib/place-element";
import {
  mouseLocationToBoundingRect,
  placeElement,
} from "@rilldata/web-common/lib/place-element";
import type { View } from "svelte-vega";
import type { VLTooltipFormatter } from "./types";

const TOOLTIP_ID = "rill-vg-tooltip";

export class VegaLiteTooltipHandler {
  location: Location = "left";
  alignment: Alignment = "middle";
  distance = 0;
  pad = 8;
  public valueFormatter: VLTooltipFormatter;

  constructor(valueFormatter: VLTooltipFormatter) {
    this.valueFormatter = valueFormatter;
  }

  handleTooltip = (
    _view: View,
    event: MouseEvent,
    _item: unknown,
    value: unknown,
  ) => {
    const existingEl = document.getElementById(TOOLTIP_ID);
    if (existingEl) {
      existingEl.remove();
    }
    if (value == null || value === "") {
      return;
    }

    const el = document.createElement("div");
    el.setAttribute("id", TOOLTIP_ID);

    const formattedValue = this.valueFormatter(value);
    el.innerHTML = formattedValue;
    document.body.appendChild(el);

    const parentBoundingClientRect = mouseLocationToBoundingRect({
      x: event.x,
      y: event.y,
    });
    const elementBoundingClientRect = el.getBoundingClientRect();

    const [leftPos, topPos] = placeElement({
      location: this.location,
      alignment: this.alignment,
      distance: this.distance,
      pad: this.pad,
      parentPosition: parentBoundingClientRect,
      elementPosition: elementBoundingClientRect,
    });

    el.setAttribute("style", `top: ${topPos}px; left: ${leftPos}px`);
  };
}
