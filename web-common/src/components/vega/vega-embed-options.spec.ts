import type { TooltipHandler } from "vega-typings";
import { afterEach, describe, expect, it } from "vitest";
import { createBaseEmbedOptions } from "./vega-embed-options";

const BOUNDARY_CLASS = "dashboard-theme-boundary";

function tooltipElement() {
  return document.getElementById("vg-tooltip-element");
}

function showAndHide(tooltip: TooltipHandler) {
  const event = new MouseEvent("pointermove", { clientX: 10, clientY: 10 });
  tooltip(null, event, null as never, { value: 1 });
  const classesWhenShown = [...tooltipElement()!.classList];
  tooltip(null, event, null as never, null);
  const classesWhenHidden = [...tooltipElement()!.classList];
  return { classesWhenShown, classesWhenHidden };
}

describe("createBaseEmbedOptions tooltip", () => {
  afterEach(() => {
    tooltipElement()?.remove();
  });

  it("passes plain tooltip options outside a theme boundary", () => {
    const options = createBaseEmbedOptions({ colorMapping: [] });
    expect(options.tooltip).toEqual({ theme: "light" });
  });

  it("adds the theme boundary class to the tooltip only while it is shown", () => {
    const options = createBaseEmbedOptions({
      colorMapping: [],
      themeBoundaryClass: BOUNDARY_CLASS,
    });
    expect(typeof options.tooltip).toBe("function");

    const { classesWhenShown, classesWhenHidden } = showAndHide(
      options.tooltip as TooltipHandler,
    );
    expect(classesWhenShown).toContain(BOUNDARY_CLASS);
    expect(classesWhenShown).toContain("light-theme");
    // Charts outside a boundary reuse the same element, so it must not keep the class.
    expect(classesWhenHidden).not.toContain(BOUNDARY_CLASS);
  });
});
