import { getNestedRowDimensionWidthKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
import { describe, expect, it } from "vitest";

describe("getNestedRowDimensionWidthKey", () => {
  it("uses the first row dimension as the stable nested row column key", () => {
    const initialKey = getNestedRowDimensionWidthKey([
      { name: "country" },
      { name: "city" },
    ]);

    expect(initialKey).toBe(
      getNestedRowDimensionWidthKey([
        { name: "country" },
        { name: "city" },
        { name: "postal_code" },
      ]),
    );
  });

  it("changes when the adjusted row dimension is removed", () => {
    expect(getNestedRowDimensionWidthKey([{ name: "country" }])).toBe(
      "country",
    );
    expect(getNestedRowDimensionWidthKey([{ name: "city" }])).toBe("city");
  });
});
