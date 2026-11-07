import {
  calculateColumnWidth,
  clampColumnWidth,
  distributeColumnWidthsToFillContainer,
  fitColumnWidthsToContainer,
  getNestedRowDimensionWidthKey,
  layoutColumnWidths,
  roleWidthBounds,
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

describe("pinned columns", () => {
  it("gives pinned columns no extra width when filling", () => {
    const widths = distributeColumnWidthsToFillContainer(
      [
        { width: 160, role: "dimension", pinned: true },
        { width: 100, role: "measure" },
        { width: 100, role: "measure" },
      ],
      560,
    );
    expect(widths[0]).toBe(160);
    expect(widths[1]).toBe(200);
    expect(widths[2]).toBe(200);
  });

  it("returns the base widths when every column is pinned", () => {
    expect(
      distributeColumnWidthsToFillContainer(
        [
          { width: 160, role: "dimension", pinned: true },
          { width: 100, role: "measure", pinned: true },
        ],
        800,
      ),
    ).toEqual([160, 100]);
  });
});

describe("layoutColumnWidths", () => {
  it("shrinks only the unpinned columns when fitting an overflow", () => {
    const widths = layoutColumnWidths(
      [
        { width: 200, role: "dimension", pinned: true },
        { width: 200, role: "measure" },
        { width: 200, role: "measure" },
      ],
      400,
      { fill: true, fit: true },
    );
    expect(widths[0]).toBe(200);
    expect(widths[1]).toBe(100);
    expect(widths[2]).toBe(100);
  });

  it("never shrinks below the role minimums", () => {
    const widths = layoutColumnWidths(
      [
        { width: 100, role: "dimension" },
        { width: 60, role: "measure" },
      ],
      100,
      { fill: true, fit: true },
    );
    expect(widths).toEqual([100, 60]);
  });

  it("stretches on underflow when filling, even with fit on", () => {
    const widths = layoutColumnWidths(
      [
        { width: 100, role: "measure" },
        { width: 100, role: "measure" },
      ],
      400,
      { fill: true, fit: true },
    );
    expect(widths).toEqual([200, 200]);
  });

  it("returns the base widths without fill or fit", () => {
    expect(
      layoutColumnWidths(
        [
          { width: 300, role: "measure" },
          { width: 300, role: "measure" },
        ],
        200,
        { fill: false, fit: false },
      ),
    ).toEqual([300, 300]);
  });
});

describe("roleWidthBounds / clampColumnWidth", () => {
  it("clamps configured widths into the resizer bounds of the role", () => {
    expect(roleWidthBounds("measure")).toEqual({ min: 60, max: 300 });
    expect(roleWidthBounds("dimension")).toEqual({ min: 100, max: 600 });
    expect(clampColumnWidth("measure", 1000)).toBe(300);
    expect(clampColumnWidth("dimension", 10)).toBe(100);
    expect(clampColumnWidth("dimension", 240.6)).toBe(241);
  });
});

describe("calculateColumnWidth", () => {
  const rows = [
    { flight_start: "2026-01-15T00:00:00Z", impressions: 10 },
    { flight_start: "2026-02-01T00:00:00Z", impressions: 20 },
  ];

  it("samples the data by column id, not by the display label", () => {
    const labelled = calculateColumnWidth("flight_start", "Start", "ts", rows);
    const unlabelled = calculateColumnWidth(
      "flight_start",
      "flight_start",
      "ts",
      rows,
    );
    // "2026-01-15T00:00:00Z" is 20 characters: 20 * 7 + 16.
    expect(labelled).toBe(156);
    expect(unlabelled).toBe(156);
  });

  it("falls back to the label when there is no data for the column", () => {
    expect(
      calculateColumnWidth("missing", "A long header label", "ts", rows),
    ).toBe(19 * 7 + 16);
  });

  it("keeps the short default for the time dimension", () => {
    expect(
      calculateColumnWidth("ts_rill_TIME_GRAIN_DAY", "Time (Day)", "ts", rows),
    ).toBe(100);
  });
});
