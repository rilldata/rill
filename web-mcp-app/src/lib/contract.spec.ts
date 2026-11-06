import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/bar_chart.json";
import empty from "../../harness/fixtures/empty.json";
import flintBar from "../../harness/fixtures/flint_bar_chart.json";
import noData from "../../harness/fixtures/no_data.json";
import unsupported from "../../harness/fixtures/unsupported_widget.json";
import {
  hasChartData,
  httpUrl,
  normalizeRows,
  parseChartDataResponse,
  parseChartPayload,
  textContent,
  withChartData,
  type ChartDataPayload,
  type ChartResult,
  type FlintChartResult,
} from "./contract";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as ChartResult;

describe("parseChartPayload", () => {
  it("accepts the canonical bar chart result", () => {
    const parsed = parseChartPayload(bar.result.structuredContent);
    expect(parsed.kind).toBe("chart");
    if (parsed.kind !== "chart") return;
    expect(parsed.result.chart_type).toBe("bar_chart");
    expect(parsed.result.title).toBe("Total Bids per Advertiser Name");
    expect(hasChartData(parsed.result)).toBe(true);
  });

  it("rejects a payload without a widget mode", () => {
    const withoutWidget = Object.fromEntries(
      Object.entries(bar.result.structuredContent).filter(
        ([key]) => key !== "widget",
      ),
    );
    expect(parseChartPayload(withoutWidget).kind).toBe("invalid");
  });

  it("reports unsupported widget modes together with the deep link", () => {
    const parsed = parseChartPayload(unsupported.result.structuredContent);
    expect(parsed).toEqual({
      kind: "unsupported_widget",
      widget: "dashboard",
      openUrl: "https://ui.rilldata.com/demo/bids/canvas/overview",
    });
  });

  it("rejects malformed payloads", () => {
    const chart = { widget: "chart", chart_type: "bar_chart", spec: {} };
    expect(parseChartPayload(undefined).kind).toBe("invalid");
    expect(parseChartPayload("text").kind).toBe("invalid");
    expect(parseChartPayload({ ...chart, chart_type: "pivot" }).kind).toBe(
      "invalid",
    );
    expect(parseChartPayload({ ...chart, spec: "x" }).kind).toBe("invalid");
    expect(parseChartPayload({ ...chart, data: "rows" }).kind).toBe("invalid");
    expect(
      parseChartPayload({ ...chart, data: { columns: [], rows: [] } }).kind,
    ).toBe("invalid");
    expect(parseChartPayload({ ...chart, fields: [] }).kind).toBe("invalid");
    expect(parseChartPayload(chart).kind).toBe("chart");
  });

  it("keeps the deep link on invalid payloads", () => {
    const parsed = parseChartPayload({
      widget: "chart",
      chart_type: "nope",
      open_url: "https://example.com/open",
    });
    expect(parsed.kind).toBe("invalid");
    if (parsed.kind === "invalid") {
      expect(parsed.openUrl).toBe("https://example.com/open");
    }
  });

  it("drops a deep link that is not http(s)", () => {
    const chart = { widget: "chart", chart_type: "bar_chart", spec: {} };
    for (const openUrl of [
      "javascript:alert(1)",
      "data:text/html,<script>alert(1)</script>",
      "/relative/path",
      42,
    ]) {
      const parsed = parseChartPayload({ ...chart, open_url: openUrl });
      expect(parsed.kind === "chart" && parsed.result.open_url).toBe(undefined);
      const invalid = parseChartPayload({
        ...chart,
        chart_type: "nope",
        open_url: openUrl,
      });
      expect(invalid.kind === "invalid" && invalid.openUrl).toBe(undefined);
    }
    const parsed = parseChartPayload({
      ...chart,
      open_url: "https://ui.rilldata.com/demo/bids/explore/bids",
    });
    expect(parsed.kind === "chart" && parsed.result.open_url).toBe(
      "https://ui.rilldata.com/demo/bids/explore/bids",
    );
    expect(httpUrl("http://localhost:9009/-/open")).toBe(
      "http://localhost:9009/-/open",
    );
  });
});

describe("parseChartDataResponse", () => {
  const flint = flintBar.result.structuredContent as FlintChartResult;
  const section = { data: flint.data, fields: flint.fields };

  it("accepts a data section, with or without the effective input", () => {
    expect(parseChartDataResponse(section)).toEqual({
      kind: "data",
      response: section,
    });
    expect(parseChartDataResponse({ data: [] }).kind).toBe("data");
    const withInput = { ...section, input: flint.input };
    expect(parseChartDataResponse(withInput)).toEqual({
      kind: "data",
      response: withInput,
    });
  });

  it("rejects a result that cannot replace the data", () => {
    for (const payload of [
      undefined,
      "rows",
      {},
      { data: "rows" },
      { data: [], fields: [] },
      { data: [], domain_values: "x" },
      { data: [], semantic_types: null },
      { data: [], input: { chart_spec: {} } },
    ]) {
      expect(
        parseChartDataResponse(payload).kind,
        JSON.stringify(payload),
      ).toBe("invalid");
    }
  });
});

describe("hasChartData", () => {
  it("is true for any data array, including an empty one", () => {
    expect(result(empty).data).toEqual([]);
    expect(hasChartData(result(empty))).toBe(true);
    expect(hasChartData({ data: [] })).toBe(true);
  });

  it("is false when the data section was not forwarded", () => {
    expect(hasChartData(result(noData))).toBe(false);
    expect(hasChartData({ fields: result(bar).fields })).toBe(false);
  });
});

describe("normalizeRows", () => {
  it("passes keyed rows through and drops non-objects", () => {
    expect(normalizeRows([{ a: 1 }, null as never, 3 as never])).toEqual([
      { a: 1 },
    ]);
    expect(normalizeRows(undefined)).toEqual([]);
  });
});

describe("withChartData", () => {
  it("replaces the data section and keeps the rest of the result", () => {
    const base = result(bar);
    const spec = { ...base.spec, time_grain: "TIME_GRAIN_WEEK" as const };
    const payload: ChartDataPayload = {
      data: [],
      fields: base.fields,
      resolved_time_range: base.resolved_time_range,
    };
    const replaced = withChartData(base, spec, payload);
    expect(replaced.spec).toBe(spec);
    expect(replaced.title).toBe(base.title);
    expect(replaced.open_url).toBe(base.open_url);
    expect(replaced.theme).toBe(base.theme);
    // A zero-row chart_data result has no `domain_values`; the previous ones must not show through.
    expect(replaced.data).toEqual([]);
    expect(replaced.domain_values).toBeUndefined();
    expect(hasChartData(replaced)).toBe(true);
  });
});

describe("textContent", () => {
  it("joins text blocks", () => {
    expect(
      textContent({
        content: [
          { type: "text", text: "a" },
          { type: "image", data: "", mimeType: "image/png" },
          { type: "text", text: "b" },
        ],
      }),
    ).toBe("a\nb");
  });
});
