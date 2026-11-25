import { describe, expect, it } from "vitest";
import { get } from "svelte/store";
import type { CanvasEntity } from "./canvas-entity";
import { Grid } from "./grid";

describe("Grid.updateFromCanvasRows", () => {
  it("keeps a row without a height apart from a row with one", () => {
    const grid = new Grid({} as CanvasEntity);

    grid.updateFromCanvasRows([
      { items: [{ component: "orders", width: 12 }] },
      { height: 100, items: [{ component: "revenue", width: 12 }] },
    ]);

    const [unset, explicit] = get(grid);
    expect(get(unset.height)).toBeUndefined();
    expect(get(explicit.height)).toBe(100);
  });

  it("clears the height when the row at a position no longer has one", () => {
    const grid = new Grid({} as CanvasEntity);

    grid.updateFromCanvasRows([
      { height: 128, items: [{ component: "totals", width: 12 }] },
      { items: [{ component: "orders", width: 12 }] },
    ]);
    // The first row is removed, so the row without a height takes its place.
    grid.updateFromCanvasRows([
      { items: [{ component: "orders", width: 12 }] },
    ]);

    const [row] = get(grid);
    expect(get(row.height)).toBeUndefined();
  });
});
