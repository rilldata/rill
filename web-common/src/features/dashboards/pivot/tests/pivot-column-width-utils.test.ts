import { getNestedRowDimensionWidthKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
import { describe, expect, it } from "vitest";

describe("getNestedRowDimensionWidthKey", () => {
  it("uses the first row dimension as the stable nested row column key", () => {
    const initialKey = getNestedRowDimensionWidthKey("explore:Sales Explore", [
      { name: "country" },
      { name: "city" },
    ]);

    expect(initialKey).toBe(
      getNestedRowDimensionWidthKey("explore:Sales Explore", [
        { name: "country" },
        { name: "city" },
        { name: "postal_code" },
      ]),
    );
  });

  it("changes when the adjusted row dimension is removed", () => {
    expect(
      getNestedRowDimensionWidthKey("explore:Sales Explore", [
        { name: "country" },
      ]),
    ).toBe("explore:Sales Explore:country");
    expect(
      getNestedRowDimensionWidthKey("explore:Sales Explore", [
        { name: "city" },
      ]),
    ).toBe("explore:Sales Explore:city");
  });

  it("scopes row dimension widths by pivot table instance", () => {
    expect(
      getNestedRowDimensionWidthKey("canvas:sales:component-1", [
        { name: "country" },
      ]),
    ).not.toBe(
      getNestedRowDimensionWidthKey("canvas:sales:component-2", [
        { name: "country" },
      ]),
    );
  });
});
