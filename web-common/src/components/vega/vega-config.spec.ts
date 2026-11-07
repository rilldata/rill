import * as vega from "vega";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import { getRillTheme } from "./vega-config";

const sequential = [
  "#e4effc",
  "#c5dbf2",
  "#98c3f0",
  "#60a3e6",
  "#2388e1",
  "#006edb",
  "#0056b8",
  "#004094",
  "#002b70",
];
const diverging = [
  "#e51029",
  "#ff643d",
  "#ffab66",
  "#fde0a5",
  "#f7f7e8",
  "#b8eeff",
  "#7bd1f9",
  "#36aaec",
  "#007be0",
  "#004cc7",
  "#1b2193",
];
const theme: Record<string, string> = {
  primary: "#4f46e5",
  ...Object.fromEntries(
    sequential.map((color, i) => [`color-sequential-${i + 1}`, color]),
  ),
  ...Object.fromEntries(
    diverging.map((color, i) => [`color-diverging-${i + 1}`, color]),
  ),
};

/** Renders a spec with Rill's theme and returns the fill of each mark item, in data order. */
async function fills(spec: TopLevelSpec): Promise<string[]> {
  const config = getRillTheme(false, theme);
  const view = new vega.View(vega.parse(compile(spec, { config }).spec), {
    renderer: "none",
  });
  await view.runAsync();
  type Node = { role?: string; items?: Node[]; fill?: string };
  const found: string[] = [];
  const visit = (node: Node) => {
    if (node.role === "mark") {
      for (const item of node.items ?? []) {
        if (item.fill) found.push(item.fill);
      }
    }
    for (const item of node.items ?? []) item.items?.forEach(visit);
  };
  visit((view.scenegraph() as unknown as { root: Node }).root);
  return found;
}

describe("getRillTheme", () => {
  it("spreads ordinal colors across the sequential palette rather than taking its faintest", async () => {
    // Ordinal color is what a time field on color gets, such as a line per month.
    const colors = await fills({
      data: {
        values: [
          { month: "2026-01" },
          { month: "2026-02" },
          { month: "2026-03" },
        ],
      },
      mark: "bar",
      encoding: {
        x: { field: "month", type: "nominal" },
        color: { field: "month", type: "ordinal" },
      },
    });

    // Sampled from the palette's darker three quarters, lightest first, where a list would
    // hand the three series its first, faintest colors.
    expect(colors).toEqual([
      "rgb(70, 150, 228)",
      "rgb(0, 110, 219)",
      "rgb(1, 75, 166)",
    ]);
  });

  it("draws color around a midpoint with Rill's diverging palette", async () => {
    // A scale with a `domainMid` (Flint's diverging color, once its own scheme is removed)
    // asks for the `diverging` range.
    const colors = await fills({
      data: {
        values: [
          { x: "a", v: -10 },
          { x: "b", v: 0 },
          { x: "c", v: 10 },
        ],
      },
      mark: "bar",
      encoding: {
        x: { field: "x", type: "nominal" },
        color: { field: "v", type: "quantitative", scale: { domainMid: 0 } },
      },
    });

    // The palette's ends, and its middle (the sixth of its eleven colors) at the midpoint.
    expect(colors).toEqual([
      "rgb(229, 16, 41)",
      "rgb(184, 238, 255)",
      "rgb(27, 33, 147)",
    ]);
  });
});
