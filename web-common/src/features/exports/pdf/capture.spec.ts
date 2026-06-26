import { describe, expect, it } from "vitest";
import { inlineSvgStyles } from "./capture";

describe("inlineSvgStyles", () => {
  it("restores original SVG style attributes", () => {
    const root = document.createElement("div");
    root.innerHTML = `
      <svg style="color: red">
        <path style="stroke-width: 2" />
        <circle />
      </svg>
    `;

    const svg = root.querySelector("svg")!;
    const path = root.querySelector("path")!;
    const circle = root.querySelector("circle")!;

    const restore = inlineSvgStyles(root);
    expect(svg.getAttribute("style")).not.toBe("color: red");
    expect(path.getAttribute("style")).not.toBe("stroke-width: 2");
    expect(circle.hasAttribute("style")).toBe(true);

    restore();
    expect(svg.getAttribute("style")).toBe("color: red");
    expect(path.getAttribute("style")).toBe("stroke-width: 2");
    expect(circle.hasAttribute("style")).toBe(false);
  });
});
