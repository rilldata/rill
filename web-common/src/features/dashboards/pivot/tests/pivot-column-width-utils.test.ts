import {
  distributeColumnWidthsToFillContainer,
  fitColumnWidthsToContainer,
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

describe("fitColumnWidthsToContainer", () => {
  const measure = (width: number) => ({ width, min: 60, max: 300 });
  const dimension = (width: number) => ({ width, min: 100, max: 600 });

  it("returns an empty list for no columns", () => {
    expect(fitColumnWidthsToContainer([], 500)).toEqual([]);
  });

  it("leaves widths unchanged when they already fit exactly", () => {
    expect(
      fitColumnWidthsToContainer([dimension(160), measure(100)], 260),
    ).toEqual([160, 100]);
  });

  it("shrinks overflowing columns in proportion to their slack above min", () => {
    // Slack: 100, 40, 140 (total 280). Deficit 140 → shrink by half the slack.
    const widths = fitColumnWidthsToContainer(
      [dimension(200), measure(100), measure(200)],
      360,
    );
    expect(widths).toEqual([150, 80, 130]);
    expect(widths.reduce((sum, w) => sum + w, 0)).toBeLessThanOrEqual(360);
  });

  it("never shrinks a column below its minimum", () => {
    const widths = fitColumnWidthsToContainer(
      [dimension(120), measure(70), measure(300)],
      300,
    );
    expect(widths[0]).toBeGreaterThanOrEqual(100);
    expect(widths[1]).toBeGreaterThanOrEqual(60);
    expect(widths[2]).toBeGreaterThanOrEqual(60);
    expect(widths.reduce((sum, w) => sum + w, 0)).toBeLessThanOrEqual(300);
  });

  it("sits every column at its minimum when even the minimums overflow", () => {
    expect(
      fitColumnWidthsToContainer(
        [dimension(200), measure(100), measure(100)],
        100,
      ),
    ).toEqual([100, 60, 60]);
  });

  it("stretches underflowing columns in proportion to their headroom below max", () => {
    // Headroom: 440, 200 (total 640). Extra 320 → grow by half the headroom.
    expect(
      fitColumnWidthsToContainer([dimension(160), measure(100)], 580),
    ).toEqual([380, 200]);
  });

  it("never stretches a column beyond its maximum", () => {
    expect(
      fitColumnWidthsToContainer([measure(100), measure(100)], 2000),
    ).toEqual([300, 300]);
  });

  it("rounds down so the total never exceeds the available width", () => {
    const widths = fitColumnWidthsToContainer(
      [measure(100), measure(100), measure(100)],
      250,
    );
    expect(widths.reduce((sum, w) => sum + w, 0)).toBeLessThanOrEqual(250);
  });
});
