import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { describe, expect, it } from "vitest";
import { sliceColumnAxesDataForDef } from "./pivot-infinite-scroll";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

function getConfig(measureNames: string[]): PivotDataStoreConfig {
  return {
    measureNames,
    rowDimensionNames: ["publisher"],
    colDimensionNames: ["domain"],
    allMeasures: [],
    allDimensions: [],
    whereFilter: createAndExpression([]),
    pivot: {
      rows: [],
      columns: [],
      sorting: [],
      expanded: {},
      columnPage: 1,
      rowPage: 1,
      enableComparison: false,
      tableMode: "nest",
      activeCell: null,
      showTotalsColumn: true,
      showTotalsRow: true,
    },
    time: {
      timeStart: undefined,
      timeEnd: undefined,
      timeZone: "UTC",
      timeDimension: "timestamp",
    },
    enableComparison: false,
    comparisonTime: undefined,
    searchText: undefined,
    isFlat: false,
  };
}

// A totals row with the grand total of the first measure,
// and one leaf column per measure under each domain value.
function getTotalsRow(domainCount: number, measureCount: number) {
  const totalsRow: PivotDataRow = { publisher: "Total", impressions: 1 };
  for (let v = 0; v < domainCount; v++) {
    for (let m = 0; m < measureCount; m++) totalsRow[`c0v${v}m${m}`] = 1;
  }
  return totalsRow;
}

describe("sliceColumnAxesDataForDef", () => {
  const domains = Array.from({ length: 30 }, (_, i) => `domain-${i}`);

  it("counts the leaf columns it keeps and the ones the data has", () => {
    // 30 domains with 3 measures each make 90 leaf columns. The first page
    // holds 50 of them, which reaches into the 17th domain, and every measure
    // of that domain is rendered.
    const result = sliceColumnAxesDataForDef(
      getConfig(["impressions", "clicks", "spend"]),
      { domain: domains },
      getTotalsRow(30, 3),
    );

    expect(result.axes.domain).toEqual(domains.slice(0, 17));
    expect(result.shownColumns).toBe(51);
    expect(result.totalColumns).toBe(90);
  });

  it("keeps every column when they fit on the page", () => {
    const result = sliceColumnAxesDataForDef(
      getConfig(["impressions"]),
      { domain: domains },
      getTotalsRow(30, 1),
    );

    expect(result.axes.domain).toEqual(domains);
    expect(result.shownColumns).toBe(30);
    expect(result.totalColumns).toBe(30);
  });
});
