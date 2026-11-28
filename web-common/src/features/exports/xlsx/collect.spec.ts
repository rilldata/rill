import {
  formatResolvedContent,
  unformatResolvedContent,
} from "@rilldata/web-common/features/canvas/components/markdown/util";
import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { readable } from "svelte/store";
import { describe, expect, it } from "vitest";
import type { MeasuredBlock } from "../pdf/capture";
import {
  buildWorkbookRequest,
  customChartTable,
  groupBySheet,
  jsonCell,
  kpiTrendTable,
  leaderboardTable,
  markdownTextBlock,
  needsImage,
  stackRows,
  toDataSheet,
} from "./collect";
import type { ExportTable } from "./payload";
import {
  V1TimeGrain,
  type V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { CAPTURE_FAILED_NOTE } from "../rasterize";

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
    isBand: false,
    ...overrides,
  };
}

describe("groupBySheet", () => {
  it("separates the canvas's own rows from each tab and drops the bands", () => {
    const { main, tabs } = groupBySheet([
      block("a", 0, 0, 0),
      block("label", 1, 0, 220, { isBand: true }),
      // A band of viewer controls has no tab attributes, so it would land on the main sheet.
      block("controls", 1, 0, 240, { isBand: true }),
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

  it("lays out the items a narrow canvas stacks in one row as rows of their own", () => {
    // The row's middle item, at 220..420, was left out.
    const positions = stackRows(
      [block("a", 0, 0, 0), block("c", 0, 0, 440), block("d", 1, 0, 660)],
      16,
    );
    expect(positions.get("a")).toStrictEqual({ x: 0, y: 0 });
    expect(positions.get("c")).toStrictEqual({ x: 0, y: 216 });
    expect(positions.get("d")).toStrictEqual({ x: 0, y: 432 });
  });
});

describe("buildWorkbookRequest", () => {
  it("leaves no gap for a markdown component that shows nothing", async () => {
    const markdown = (id: string, text: string) =>
      ({
        id,
        type: "markdown",
        exportPayload: readable({ kind: "text", text, unformattedText: text }),
        specStore: readable({}),
      }) as unknown as BaseCanvasComponent;
    const components = new Map([
      ["intro", markdown("intro", "Intro")],
      ["styles", markdown("styles", "<style>.card { color: red }</style>")],
      ["outro", markdown("outro", "Outro")],
    ]);
    const request = await buildWorkbookRequest({
      canvasName: "canvas",
      canvasEntity: {
        componentsStore: readable(components),
        layout: readable([]),
      } as unknown as CanvasEntity,
      blocks: [
        block("intro", 0, 0, 0),
        block("styles", 1, 0, 220),
        block("outro", 2, 0, 440),
      ],
      images: new Map(),
      chartMode: "native",
      header: {
        title: "Canvas",
        timeRangeText: "",
        comparisonText: "",
        timeZone: "UTC",
        filterLines: [],
      },
      generatedAt: "",
      dashboardUrl: "",
    });
    expect(
      request.dashboardSheets[0].blocks.map((b) => [
        b.component,
        b.geometry?.yPx,
      ]),
    ).toStrictEqual([
      ["intro", 0],
      ["outro", 240],
    ]);
  });

  // A custom chart is exported as a snapshot. Without one (its capture failed)
  // the block is a note, so the runtime is never sent an image it would reject
  // and the rest of the workbook still exports.
  it("notes a block whose snapshot failed and keeps the others", async () => {
    const customChart = (id: string) =>
      ({
        id,
        type: "custom_chart",
        exportPayload: readable({ kind: "customChart", results: [] }),
        specStore: readable({ title: id }),
      }) as unknown as BaseCanvasComponent;
    const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47]);

    const request = await buildWorkbookRequest({
      canvasName: "canvas",
      canvasEntity: {
        componentsStore: readable(
          new Map([
            ["captured", customChart("captured")],
            ["failed", customChart("failed")],
          ]),
        ),
        layout: readable([]),
      } as unknown as CanvasEntity,
      blocks: [block("captured", 0, 0, 0), block("failed", 1, 0, 220)],
      images: new Map([["captured", png]]),
      chartMode: "native",
      header: {
        title: "Canvas",
        timeRangeText: "",
        comparisonText: "",
        timeZone: "UTC",
        filterLines: [],
      },
      generatedAt: "2026-01-01 00:00 UTC",
      dashboardUrl: "https://example.com/canvas",
    });

    const [captured, failed] = request.dashboardSheets[0].blocks;
    expect(captured.content).toMatchObject({
      case: "image",
      value: { png },
    });
    expect(failed.content).toMatchObject({
      case: "note",
      value: { text: CAPTURE_FAILED_NOTE },
    });
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

describe("markdownTextBlock", () => {
  // The block of a markdown text, with the data sheets it adds.
  function textBlock(text: string, unformattedText = text) {
    const sheets: { suffix: string; title: string; table: ExportTable }[] = [];
    const block = markdownTextBlock(
      { kind: "text", text, unformattedText },
      (suffix, title, table) => {
        sheets.push({ suffix, title, table });
        return `sheet-${sheets.length}`;
      },
    );
    return { block: block?.toJson(), sheets };
  }

  it("is plain text for markdown and HTML without a table", () => {
    expect(
      textBlock(
        `# Title\n\n<div class="card"><b>Revenue</b><div>$1.2M</div></div>`,
      ),
    ).toStrictEqual({
      block: { text: "Title\nRevenue\n$1.2M" },
      sheets: [],
    });
  });

  it("is nothing for a text that shows nothing", () => {
    expect(
      textBlock(`<style>.card { color: red }</style>`).block,
    ).toBeUndefined();
  });

  it("puts each table on a data sheet and keeps the text around it, in reading order", () => {
    const { block, sheets } = textBlock(`## Scorecard

<table>
  <tr><th>Metric</th><th>Actual</th></tr>
  <tr><td>Spend</td><td>$1,200</td></tr>
  <tr><td>Win rate</td><td>25%</td></tr>
</table>

Targets are set per quarter.

| Rep | Deals |
|---|---|
| Ana | 12 |
`);
    expect(block).toStrictEqual({
      parts: [
        { text: "Scorecard" },
        { dataSheetId: "sheet-1" },
        { text: "Targets are set per quarter." },
        { dataSheetId: "sheet-2" },
      ],
    });
    // Both tables sit under the same heading; the suffix tells their sheets apart.
    expect(sheets.map(({ suffix, title }) => [suffix, title])).toStrictEqual([
      ["table 1", "Scorecard"],
      ["table 2", "Scorecard"],
    ]);
    expect(sheets[0].table).toStrictEqual({
      columns: [
        {
          name: "col_1",
          displayName: "Metric",
          type: WorkbookColumnType.STRING,
          numberFormat: undefined,
        },
        {
          name: "col_2",
          displayName: "Actual",
          type: WorkbookColumnType.NUMBER,
          numberFormat: undefined,
        },
      ],
      rows: [
        ["Spend", 1200],
        ["Win rate", 0.25],
      ],
      rowNumberFormats: ['"$"#,##0', "0%"],
    });
    expect(sheets[1].table.rows).toStrictEqual([["Ana", 12]]);
  });

  it("writes the value behind a formatted measure with the format the text shows", () => {
    // What the runtime resolves with `apply_formatting` on, and what the component shows for it.
    const token = (field: string, value: number) =>
      `__RILL__FORMAT__(${JSON.stringify({ metrics_view: "mv", field, value })})`;
    const resolved = `| Rep | Revenue | Revenue (EUR) |\n|---|---|---|\n| Ana | ${token("revenue", 1234567.891)} | ${token("revenue_eur", 1234)} |\n`;
    const metricsView: V1MetricsView = {
      state: {
        validSpec: {
          measures: [
            { name: "revenue", formatPreset: "currency_usd" },
            // A locale that groups thousands with ".", so 1234 shows as "1.234 €".
            {
              name: "revenue_eur",
              formatD3: "$,.0f",
              formatD3Locale: {
                decimal: ",",
                thousands: ".",
                grouping: [3],
                currency: ["", "\u00a0€"],
              },
            },
          ],
        },
      },
    };
    const { block, sheets } = textBlock(
      formatResolvedContent(resolved, { mv: metricsView }),
      unformatResolvedContent(resolved),
    );
    expect(block).toStrictEqual({ parts: [{ dataSheetId: "sheet-1" }] });
    // One table needs no suffix, and without a caption or heading it is named like the component.
    expect(sheets.map(({ suffix, title }) => [suffix, title])).toStrictEqual([
      ["", ""],
    ]);
    expect(sheets[0].table.rows).toStrictEqual([["Ana", 1234567.891, 1234]]);
    expect(
      sheets[0].table.columns.map((column) => column.numberFormat),
    ).toStrictEqual([undefined, '"$"#,##0.00', '#,##0"€"']);
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
      totalsLast: true,
      rowNumberFormats: ["0.0%"],
      moreRows: true,
    });
    expect(sheet.columns.map((c) => c.name)).toStrictEqual(["week", "spend"]);
    expect(sheet.rows[0].toJson()).toStrictEqual(["w1", 1]);
    expect(sheet.totals?.toJson()).toStrictEqual(["Total", 1]);
    expect(sheet.totalsLast).toBe(true);
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
    expect(
      needsImage(
        component,
        { kind: "text", text: "x", unformattedText: "x" },
        "native",
      ),
    ).toBe(false);
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
