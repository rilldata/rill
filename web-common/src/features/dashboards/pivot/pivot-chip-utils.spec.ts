import {
  dedupePivotChips,
  insertPivotChipAt,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-chip-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { describe, expect, it } from "vitest";

describe("dedupePivotChips", () => {
  const dim = (id: string): PivotChipData => ({
    id,
    title: id,
    type: PivotChipType.Dimension,
  });

  it("keeps the first occurrence of a repeated id", () => {
    expect(dedupePivotChips([dim("a"), dim("b"), dim("a")])).toEqual([
      dim("a"),
      dim("b"),
    ]);
  });

  it("drops chips with an empty id", () => {
    expect(dedupePivotChips([dim("a"), dim(""), dim("b")])).toEqual([
      dim("a"),
      dim("b"),
    ]);
  });

  it("drops chips whose id is excluded", () => {
    expect(dedupePivotChips([dim("a"), dim("b"), dim("c")], ["b"])).toEqual([
      dim("a"),
      dim("c"),
    ]);
  });
});

describe("insertPivotChipAt", () => {
  const dim = (id: string): PivotChipData => ({
    id,
    title: id,
    type: PivotChipType.Dimension,
  });
  const items = [dim("a"), dim("b"), dim("c")];

  it("inserts a chip that is not in the list", () => {
    expect(insertPivotChipAt(items, dim("d"), 1)).toEqual([
      dim("a"),
      dim("d"),
      dim("b"),
      dim("c"),
    ]);
    expect(insertPivotChipAt(items, dim("d"), 3)).toEqual([...items, dim("d")]);
  });

  it("moves a chip that is still in the list instead of duplicating it", () => {
    // The index counts the copy that is removed.
    expect(insertPivotChipAt(items, dim("a"), 3)).toEqual([
      dim("b"),
      dim("c"),
      dim("a"),
    ]);
    expect(insertPivotChipAt(items, dim("c"), 0)).toEqual([
      dim("c"),
      dim("a"),
      dim("b"),
    ]);
    expect(insertPivotChipAt(items, dim("b"), 1)).toEqual(items);
  });

  it("removes every existing copy of the chip", () => {
    expect(
      insertPivotChipAt([dim("a"), dim("b"), dim("a")], dim("a"), 3),
    ).toEqual([dim("b"), dim("a")]);
  });
});
