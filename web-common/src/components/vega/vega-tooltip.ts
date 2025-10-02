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
const VEGA_TOOLTIP_ID = "vg-tooltip-element";

export class VegaLiteTooltipHandler {
  location: Location = "left";
  alignment: Alignment = "middle";
  distance = 0;
  pad = 8;
  public valueFormatter: VLTooltipFormatter;
  private mouseLeaveHandler: ((event: Event) => void) | null = null;

  constructor(valueFormatter: VLTooltipFormatter) {
    this.valueFormatter = valueFormatter;
  }

  /**
   * There is an element for vega tooltip that does not have a position.
   * So it can add to the height, especially when dynamic height is set.
   */
  public static resetElement() {
    const el = document.getElementById(VEGA_TOOLTIP_ID);
    if (!el) return false;
    el.setAttribute("style", `top:0;left:0;`);
    return true;
  }

  removeTooltip = () => {
    const existingEl = document.getElementById(TOOLTIP_ID);
    if (existingEl) {
      existingEl.remove();
    }
  };

  handleTooltip = (
    _view: View,
    event: MouseEvent,
    _item: unknown,
    value: unknown,
  ) => {
    this.removeTooltip();

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

  removeMouseLeaveHandler() {
    if (this.mouseLeaveHandler) {
      // Find all potential containers and remove listener
      document.querySelectorAll(".vega-embed").forEach((container) => {
        container.removeEventListener("mouseleave", this.mouseLeaveHandler!);
      });
      this.mouseLeaveHandler = null;
    }
  }

  destroy() {
    this.removeTooltip();
    this.removeMouseLeaveHandler();
  }
}
