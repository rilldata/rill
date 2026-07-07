import {
  distributeColumnWidthsToFillContainer,
  getNestedRowDimensionWidthKey,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
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

describe("distributeColumnWidthsToFillContainer", () => {
  it("leaves widths unchanged when the columns already fill the container", () => {
    expect(
      distributeColumnWidthsToFillContainer(
        [
          { width: 160, role: "dimension" },
          { width: 100, role: "measure" },
        ],
        200,
      ),
    ).toEqual([160, 100]);
  });

  it("gives dimension columns more of the extra width than measure columns", () => {
    const widths = distributeColumnWidthsToFillContainer(
      [
        { width: 160, role: "dimension" },
        { width: 100, role: "measure" },
        { width: 100, role: "measure" },
      ],
      530,
    );

    expect(widths.reduce((sum, width) => sum + width, 0)).toBeCloseTo(530);
    expect(widths[0] - 160).toBeGreaterThan(widths[1] - 100);
    expect(widths[1]).toBeCloseTo(widths[2]);
  });

  it("spreads extra width evenly when there are no measure columns", () => {
    expect(
      distributeColumnWidthsToFillContainer(
        [
          { width: 160, role: "dimension" },
          { width: 120, role: "dimension" },
        ],
        380,
      ),
    ).toEqual([210, 170]);
  });
});
