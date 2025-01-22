import { Grid } from "./grid";
import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
import * as defaults from "./constants";
import { beforeEach, describe, expect, it } from "vitest";

// FIXME: re-enable this test
describe.skip("Grid", () => {
  let items: V1CanvasItem[];

  beforeEach(() => {
    items = [
      { x: 0, y: 0, width: 3, height: 2 },
      { x: 3, y: 0, width: 3, height: 2 },
      { x: 6, y: 0, width: 3, height: 2 },
      { x: 9, y: 0, width: 3, height: 2 },
    ];
  });

  describe("moveItem", () => {
    it("should handle right drop correctly", () => {
      const grid = new Grid(items);
      const result = grid.moveItem(
        items[2], // Moving the bottom item
        items[0], // Target is first item
        "right",
        2,
      );

      expect(result.items[1]).toEqual(
        expect.objectContaining({
          x: items[0].width,
          y: items[0].y,
          width: 6,
          height: 2,
        }),
      );
    });

    it("should handle left drop correctly", () => {
      const grid = new Grid(items);
      const result = grid.moveItem(
        items[1], // Moving second item
        items[0], // Target is first item
        "left",
        1,
      );

      expect(result.items[0]).toEqual(
        expect.objectContaining({
          x: 0,
          y: 0,
          width: 3,
          height: 2,
        }),
      );
    });

    it("should handle bottom drop correctly", () => {
      const grid = new Grid(items);
      const result = grid.moveItem(
        items[1], // Moving second item
        items[0], // Target is first item
        "bottom",
        1,
      );

      expect(result.items[1]).toEqual(
        expect.objectContaining({
          x: 0,
          y: items[0].height,
          width: defaults.COLUMN_COUNT,
          height: 2,
        }),
      );
    });

    it("should handle row drop correctly", () => {
      const grid = new Grid(items);
      const result = grid.moveItem(
        items[2], // Moving bottom item
        items[0], // Target is first item
        "row",
        2,
      );

      expect(result.items[0]).toEqual(
        expect.objectContaining({
          x: 0,
          y: 0,
          height: 2,
        }),
      );
    });
  });

  describe("getDropPosition", () => {
    it("should detect bottom drop zone", () => {
      const grid = new Grid(items);
      const rect = new DOMRect(0, 0, 100, 100);
      const position = grid.getDropPosition(50, 90, rect);

      expect(position).toBe("bottom");
    });

    it("should detect row drop zone", () => {
      const grid = new Grid(items);
      const rect = new DOMRect(0, 0, 100, 100);
      const position = grid.getDropPosition(50, 10, rect);

      expect(position).toBe("row");
    });

    it("should detect left drop zone", () => {
      const grid = new Grid(items);
      const rect = new DOMRect(0, 0, 100, 100);
      const position = grid.getDropPosition(20, 50, rect);

      expect(position).toBe("left");
    });

    it("should detect right drop zone", () => {
      const grid = new Grid(items);
      const rect = new DOMRect(0, 0, 100, 100);
      const position = grid.getDropPosition(80, 50, rect);

      expect(position).toBe("right");
    });
  });
});
