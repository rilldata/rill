import { describe, expect, it } from "vitest";
import type { CanvasEntity } from "./canvas-entity";
import {
  TabGroup,
  rekeyedTabGroupNames,
  rowIndexAfterMove,
  type LayoutBlock,
} from "./tab-group";

// TabGroup only stores the entity reference; the re-key helpers never touch it.
const canvas = null as unknown as CanvasEntity;

function groupBlock(rowIndex: number, name: string): LayoutBlock {
  return { kind: "tab-group", rowIndex, group: new TabGroup(canvas, name) };
}

function rowBlock(rowIndex: number, freeRowIndex: number): LayoutBlock {
  return { kind: "row", rowIndex, freeRowIndex };
}

describe("rowIndexAfterMove", () => {
  it("maps rows when a block moves up past its neighbours", () => {
    const map = rowIndexAfterMove(2, 0);
    expect([0, 1, 2, 3].map(map)).toEqual([1, 2, 0, 3]);
  });

  it("maps rows when a block moves down past its neighbours", () => {
    const map = rowIndexAfterMove(0, 2);
    expect([0, 1, 2, 3].map(map)).toEqual([2, 0, 1, 3]);
  });
});

describe("rekeyedTabGroupNames", () => {
  it("re-keys every index-named group whose row shifts, not just the moved one", () => {
    // Two unnamed groups: dragging the second above the first swaps their keys.
    const blocks = [groupBlock(0, "group-0"), groupBlock(1, "group-1")];
    expect(rekeyedTabGroupNames(blocks, rowIndexAfterMove(1, 0))).toEqual([
      ["group-0", "group-1"],
      ["group-1", "group-0"],
    ]);
  });

  it("keeps explicit names and skips plain rows and removed groups", () => {
    const blocks = [
      rowBlock(0, 0),
      groupBlock(1, "deep_dive"),
      groupBlock(2, "group-2"),
      groupBlock(3, "group-3"),
    ];
    // Delete row 2: the named group is untouched, group-3 slides up, group-2 is gone.
    const afterDelete = (rowIndex: number) =>
      rowIndex === 2 ? -1 : rowIndex > 2 ? rowIndex - 1 : rowIndex;
    expect(rekeyedTabGroupNames(blocks, afterDelete)).toEqual([
      ["deep_dive", "deep_dive"],
      ["group-3", "group-2"],
    ]);
  });

  it("shifts the groups after a duplicated block down by one", () => {
    const blocks = [groupBlock(0, "group-0"), groupBlock(1, "group-1")];
    const afterDuplicateOfFirst = (rowIndex: number) =>
      rowIndex > 0 ? rowIndex + 1 : rowIndex;
    expect(rekeyedTabGroupNames(blocks, afterDuplicateOfFirst)).toEqual([
      ["group-0", "group-0"],
      ["group-1", "group-2"],
    ]);
  });
});
