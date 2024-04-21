import {
  mouseLocationToBoundingRect,
  placeElement,
} from "@rilldata/web-common/lib/place-element";
import { VLTooltipFormatter } from "../types";

const TOOLTIP_ID = "vg-tooltip-element";

export class VegaLiteTooltipHandler {
  location = "bottom";
  alignment = "middle";
  distance = 0;
  pad = 8;
  private valueFormatter: VLTooltipFormatter;

  constructor(valueFormatter) {
    this.valueFormatter = valueFormatter;
  }

  removeTooltip() {
    const el = document.getElementById(TOOLTIP_ID);
    if (el) {
      el.remove();
    }
  }

  handleTooltip(view, event, item, value) {
    this.removeTooltip();

    // Hide tooltip for null, undefined, or empty string values
    if (value == null || value === "") {
      return;
    }

    const el = document.createElement("div");
    el.setAttribute("id", TOOLTIP_ID);

    const formattedValue = this.valueFormatter(value);
    console.log(formattedValue); // Display formatted value to console
    el.innerHTML = formattedValue; // Set inner HTML of the tooltip element

    // Add to DOM to calculate tooltip size
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
  }
}
