import { VirtualizedTableColumnSizes } from "@rilldata/web-common/components/virtualized-table/columnSizes";
import { describe, it, expect } from "vitest";

describe("VirtualizedTableColumnSizes", () => {
  it("Sanity check", () => {
    const sizes = new VirtualizedTableColumnSizes();
    expect(
      sizes.get(
        "model",
        [{ name: "col1" }, { name: "col2" }, { name: "col3" }],
        "name",
        () => [50, 50, 50],
      ),
    ).toEqual([50, 50, 50]);

    // set the width for col2 to something specific
    sizes.set("model", "col2", 100);
    // the width is saved
    expect(
      sizes.get(
        "model",
        [{ name: "col1" }, { name: "col2" }, { name: "col3" }],
        "name",
        () => [50, 50, 50],
      ),
    ).toEqual([50, 100, 50]);

    // the width is saved when one of the columns is removed
    expect(
      sizes.get("model", [{ name: "col1" }, { name: "col2" }], "name", () => [
        50, 50,
      ]),
    ).toEqual([50, 100]);

    expect(
      sizes.get("model", [{ name: "col1" }, { name: "col4" }], "name", () => [
        50, 50,
      ]),
    ).toEqual([50, 50]);
    // removing a column will remove any saved sizes. this is for cleaning up leaks
    expect(
      sizes.get(
        "model",
        [{ name: "col1" }, { name: "col2" }, { name: "col4" }],
        "name",
        () => [50, 50, 50],
      ),
    ).toEqual([50, 50, 50]);
  });
});
