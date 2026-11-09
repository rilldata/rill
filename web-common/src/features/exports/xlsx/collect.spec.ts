import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { describe, expect, it } from "vitest";
import type { MeasuredBlock } from "../pdf/capture";
import {
  customChartTable,
  groupBySheet,
  jsonCell,
  kpiTrendTable,
  leaderboardTable,
  needsImage,
  stackRows,
  stripMarkdown,
  toDataSheet,
} from "./collect";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";

function block(
  id: string,
  rowIndex: number,
  xPx: number,
  yPx: number,
  overrides: Partial<MeasuredBlock> = {},
): MeasuredBlock {
  return {
    id,
    element: null as unknown as HTMLElement,
    xPx,
    yPx,
    widthPx: 400,
    heightPx: 200,
    rowIndex,
    isTabLabel: false,
    ...overrides,
  };
}

describe("groupBySheet", () => {
  it("separates the canvas's own rows from each tab and drops label bands", () => {
    const { main, tabs } = groupBySheet([
      block("a", 0, 0, 0),
      block("label", 1, 0, 220, { isTabLabel: true }),
      block("b", 1, 0, 260, { tabGroup: "g", tab: "overview" }),
      block("c", 2, 0, 480, { tabGroup: "g", tab: "channels" }),
      block("d", 3, 0, 700),
    ]);
    expect(main.map((b) => b.id)).toStrictEqual(["a", "d"]);
    expect([...tabs.keys()]).toStrictEqual([
      "g\u0000overview",
      "g\u0000channels",
    ]);
    expect(tabs.get("g\u0000channels")?.map((b) => b.id)).toStrictEqual(["c"]);
  });
});

describe("stackRows", () => {
  it("closes the gaps between a sheet's rows while keeping offsets within a row", () => {
    const positions = stackRows(
      [
        block("a", 0, 0, 0),
        block("b", 0, 420, 10, { heightPx: 300 }),
        block("c", 3, 0, 900),
      ],
      16,
    );
    expect(positions.get("a")).toStrictEqual({ x: 0, y: 0 });
    expect(positions.get("b")).toStrictEqual({ x: 420, y: 10 });
    // Row 0 spans 0..310, so row 3 starts after it plus the gap.
    expect(positions.get("c")).toStrictEqual({ x: 0, y: 326 });
  });
});

describe("leaderboardTable", () => {
  it("adds comparison columns only for measures with a previous value", () => {
    const table = leaderboardTable({
      dimension: { name: "country", displayName: "Country" },
      measures: [
        {
          name: "impressions",
          displayName: "Impressions",
          formatPreset: "humanize",
        },
        { name: "ctr", displayName: "CTR", formatPreset: "percentage" },
      ],
      rows: [
        {
          dimensionValue: "US",
          uri: null,
          values: { impressions: 10, ctr: 0.5 },
          pctOfTotals: {},
          prevValues: { impressions: 8, ctr: null },
          deltaRels: { impressions: 0.25, ctr: null },
          deltaAbs: { impressions: 2, ctr: null },
          selectedIndex: -1,
        },
      ],
    });
    expect(table.columns.map((c) => c.displayName)).toStrictEqual([
      "Country",
      "Impressions",
      "CTR",
      "Impressions (previous)",
      "Impressions change",
      "Impressions % change",
    ]);
    expect(table.rows).toStrictEqual([["US", 10, 0.5, 8, 2, 0.25]]);
  });
});

describe("customChartTable", () => {
  it("types columns from the query schema and localizes timestamps", () => {
    const table = customChartTable(
      {
        data: [{ day: "2026-01-01T00:00:00Z", n: 3, ok: true, label: "x" }],
        schema: [
          { name: "day", type: "CODE_TIMESTAMP" },
          { name: "n", type: "CODE_INT64" },
          { name: "ok", type: "CODE_BOOL" },
          { name: "label", type: "CODE_STRING" },
        ],
      },
      "Asia/Kolkata",
    );
    expect(table?.columns.map((c) => c.type)).toStrictEqual([
      WorkbookColumnType.TIMESTAMP,
      WorkbookColumnType.NUMBER,
      WorkbookColumnType.BOOL,
      WorkbookColumnType.STRING,
    ]);
    expect(table?.rows).toStrictEqual([["2026-01-01T05:30:00", 3, true, "x"]]);
    expect(customChartTable({ data: [], schema: [] }, "UTC")).toBeUndefined();
  });
});

describe("jsonCell", () => {
  it("keeps JSON scalars, serializes dates and drops non-finite numbers", () => {
    expect(jsonCell(1.5)).toBe(1.5);
    expect(jsonCell(Number.NaN)).toBeNull();
    expect(jsonCell("a")).toBe("a");
    expect(jsonCell(true)).toBe(true);
    expect(jsonCell(undefined)).toBeNull();
    expect(jsonCell(new Date(Date.UTC(2026, 0, 1)))).toBe(
      "2026-01-01T00:00:00.000Z",
    );
    expect(jsonCell({ a: 1 })).toBe('{"a":1}');
  });
});

describe("stripMarkdown", () => {
  it("reduces markdown to plain text", () => {
    expect(
      stripMarkdown(
        "# Title\n\nSome **bold** and _italic_ text with a [link](https://x.y) and `code`.\n- item\n> quote",
      ),
    ).toBe(
      "Title\n\nSome bold and italic text with a link and code.\n• item\nquote",
    );
  });

  it("keeps underscores and asterisks that are not emphasis", () => {
    expect(stripMarkdown("snake_case_name costs 2 * 3 * 4, *really*")).toBe(
      "snake_case_name costs 2 * 3 * 4, really",
    );
  });

  it("drops raw HTML and decodes entities", () => {
    expect(
      stripMarkdown(
        "Spend <b>up</b> 5% &amp; CPA &lt; $2\n\n<div>hidden block</div>\n\nDone",
      ),
    ).toBe("Spend up 5% & CPA < $2\n\nDone");
  });

  it("leaves text that only looks like a tag as literal cell text", () => {
    // The parser sees one tag inside the broken one; what remains is plain
    // text in a cell, which nothing renders as HTML.
    expect(stripMarkdown("<scr<script>ipt>alert(1)</script>")).toBe(
      "<script>alert(1)",
    );
  });

  it("keeps lists, tables, code and line breaks readable", () => {
    expect(
      stripMarkdown(
        "1. one\n2. two\n\n| Channel | Spend |\n|---|---|\n| Search | **10** |\n\n```\nselect 1\n```\n\n![Logo](logo.png) first  \nsecond `a < b`",
      ),
    ).toBe(
      "1. one\n2. two\n\nChannel | Spend\nSearch | 10\n\nselect 1\n\nLogo first\nsecond a < b",
    );
  });
});

describe("toDataSheet", () => {
  it("drops columns named like an earlier one, with their cells, and keeps the more-rows flag", () => {
    const sheet = toDataSheet("s", "Spend", {
      columns: [
        { name: "week", displayName: "Week", type: WorkbookColumnType.STRING },
        {
          name: "spend",
          displayName: "Spend",
          type: WorkbookColumnType.NUMBER,
        },
        {
          name: "spend",
          displayName: "Spend (size)",
          type: WorkbookColumnType.NUMBER,
        },
      ],
      rows: [["w1", 1, 1]],
      totals: ["Total", 1, 1],
      rowNumberFormats: ["0.0%"],
      moreRows: true,
    });
    expect(sheet.columns.map((c) => c.name)).toStrictEqual(["week", "spend"]);
    expect(sheet.rows[0].toJson()).toStrictEqual(["w1", 1]);
    expect(sheet.totals?.toJson()).toStrictEqual(["Total", 1]);
    expect(sheet.rowNumberFormats).toStrictEqual(["0.0%"]);
    expect(sheet.moreRows).toBe(true);
  });
});

describe("needsImage", () => {
  it("snapshots images, maps and custom charts, and charts only where Excel has no equivalent", () => {
    const component = { type: "image" } as BaseCanvasComponent;
    expect(
      needsImage(component, { kind: "imageUrl", url: "logo.png" }, "native"),
    ).toBe(true);
    expect(needsImage(component, { kind: "text", text: "x" }, "native")).toBe(
      false,
    );
    // A static component without a payload would be captured as a blank card.
    expect(needsImage(component, undefined, "native")).toBe(false);
  });
});

describe("kpiTrendTable", () => {
  it("joins the KPIs' time series on time, one column per KPI with a trend", () => {
    const table = kpiTrendTable(
      [
        {
          measure: {
            name: "impressions",
            displayName: "Impressions",
            formatPreset: "humanize",
          },
          label: "Impressions",
          value: 10,
          comparisonLabel: undefined,
          comparisonValue: null,
          delta: null,
          deltaPercent: null,
          timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
          trend: [
            { time: "2026-01-02T00:00:00Z", value: 12 },
            { time: "2026-01-01T00:00:00Z", value: 10 },
          ],
        },
        {
          measure: undefined,
          label: "CTR",
          value: null,
          comparisonLabel: undefined,
          comparisonValue: null,
          delta: null,
          deltaPercent: null,
        },
        {
          measure: {
            name: "spend",
            displayName: "Spend",
            formatPreset: "currency_usd",
          },
          label: "Spend",
          value: 3,
          comparisonLabel: undefined,
          comparisonValue: null,
          delta: null,
          deltaPercent: null,
          timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
          trend: [{ time: "2026-01-01T00:00:00Z", value: 1.5 }],
        },
      ],
      "UTC",
    );
    expect(
      table?.columns.map((c) => [
        c.name,
        c.displayName,
        c.type,
        c.numberFormat,
      ]),
    ).toStrictEqual([
      ["time", "Time", WorkbookColumnType.DATE, "yyyy-mm-dd"],
      ["kpi_0", "Impressions", WorkbookColumnType.NUMBER, "#,##0"],
      ["kpi_2", "Spend", WorkbookColumnType.NUMBER, '"$"#,##0.00'],
    ]);
    expect(table?.rows).toStrictEqual([
      ["2026-01-01T00:00:00", 10, 1.5],
      ["2026-01-02T00:00:00", 12, null],
    ]);
  });

  it("returns nothing when no KPI has a trend", () => {
    expect(
      kpiTrendTable(
        [
          {
            measure: undefined,
            label: "x",
            value: 1,
            comparisonLabel: undefined,
            comparisonValue: null,
            delta: null,
            deltaPercent: null,
          },
        ],
        "UTC",
      ),
    ).toBeUndefined();
  });
});
