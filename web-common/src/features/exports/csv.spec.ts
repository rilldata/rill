import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import { unformatResolvedContent } from "@rilldata/web-common/features/canvas/components/markdown/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts/types";
import { componentCsv } from "@rilldata/web-common/features/exports/csv";
import type { ComponentExportPayload } from "@rilldata/web-common/features/exports/xlsx/payload";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const text = (markdown: string): ComponentExportPayload => ({
  kind: "text",
  text: markdown,
  unformattedText: unformatResolvedContent(markdown),
});

const token = (value: unknown) =>
  `__RILL__FORMAT__(${JSON.stringify({ metrics_view: "mv", field: "revenue", value })})`;

describe("componentCsv", () => {
  it("writes the rows behind a chart, with dates at a date grain and one column per field", () => {
    const payload: ComponentExportPayload = {
      kind: "chart",
      chartType: "combo_chart",
      // The same measure on both axes is one column.
      spec: {
        metrics_view: "mv",
        x: { field: "timestamp", type: "temporal" },
        y1: { field: "impressions", type: "quantitative", mark: "bar" },
        y2: { field: "impressions", type: "quantitative", mark: "line" },
      } as CanvasChartSpec,
      data: {
        data: [
          { timestamp: "2024-01-03T05:00:00Z", impressions: 1234.5 },
          { timestamp: "2024-01-02T05:00:00Z", impressions: null },
        ],
        isFetching: false,
        fields: {
          timestamp: { field: "timestamp", displayName: "Time" },
          impressions: {
            name: "impressions",
            displayName: "Impressions",
            formatPreset: "humanize",
            expression: "count(*)",
          },
        },
      } as unknown as ChartDataResult,
    };

    expect(
      componentCsv(payload, V1TimeGrain.TIME_GRAIN_DAY, "America/New_York"),
    ).toBe(`Time,Impressions
2024-01-02,
2024-01-03,1234.5
`);
  });

  it("writes each query of a custom chart, separated by an empty line", () => {
    const payload: ComponentExportPayload = {
      kind: "customChart",
      results: [
        {
          data: [
            { ts: "2024-01-02T05:30:00Z", publisher: "Acme, Inc.", bids: 7 },
          ],
          schema: [
            { name: "ts", type: "CODE_TIMESTAMP" },
            { name: "publisher", type: "CODE_STRING" },
            { name: "bids", type: "CODE_INT64" },
          ],
        },
        // A query that has not reported a schema has no table.
        { data: [], schema: [] },
        {
          data: [{ total: 0.25 }],
          schema: [{ name: "total", type: "CODE_FLOAT64" }],
        },
      ],
    };

    expect(componentCsv(payload, undefined, "America/New_York")).toBe(
      `ts,publisher,bids
2024-01-02T00:30:00,"Acme, Inc.",7

total
0.25
`,
    );
  });

  it("writes a resolved markdown table with raw measure values and plain cell text", () => {
    const markdown = `# Revenue by rep

| Rep | Region | Revenue |
|----------|--------|---------|
| **Ana** | North, East | ${token(1234567.891)} |
| [Bo "B"](https://example.com) | \`South\` | ${token(null)} |
`;

    expect(componentCsv(text(markdown), undefined, "UTC")).toBe(
      `Rep,Region,Revenue
Ana,"North, East",1234567.891
"Bo ""B""",South,
`,
    );
  });

  it("includes the HTML tables of a markdown text, one after the other", () => {
    const markdown = `| a | b |
|---|---|
| 1 | 2 |

<table>
  <tr><th>c</th></tr>
  <tr><td>
    3 &amp; 4
  </td></tr>
</table>
`;

    expect(componentCsv(text(markdown), undefined, "UTC")).toBe(`a,b
1,2

c
3 & 4
`);
  });

  it("has nothing to download for a markdown text without a table, or a component it does not cover", () => {
    expect(
      componentCsv(text("# Title\n\nSome | text"), undefined, "UTC"),
    ).toBeUndefined();
    expect(
      componentCsv({ kind: "error", message: "boom" }, undefined, "UTC"),
    ).toBeUndefined();
    expect(componentCsv(undefined, undefined, "UTC")).toBeUndefined();
  });
});
