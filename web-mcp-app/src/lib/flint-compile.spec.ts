import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import { compileFlintSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
import {
  createMeasureValueFormatter,
  humanizeDataType,
} from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { expressionFunction, parse, View } from "vega";
import { expressionInterpreter } from "vega-interpreter";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import comparison from "../../harness/fixtures/flint_line_comparison.json";
import { adaptFlintResult, type AdaptedFlintChart } from "./adapt";
import type { FlintChartResult } from "./contract";

/**
 * End-to-end check of the Flint render path per fixture: result -> adapter -> Flint compile ->
 * Vega-Lite compile -> headless Vega run with the CSP-safe interpreter, no renderer. This is the
 * wiring FlintSurface performs, minus the Svelte shell.
 */

interface Fixture {
  name: string;
  expect: string;
  result: { structuredContent?: unknown };
}

const flintFixtures = Object.entries(
  import.meta.glob<Fixture>("../../harness/fixtures/flint_*.json", {
    eager: true,
    import: "default",
  }),
).map(([path, fixture]) => ({
  file: path.split("/").pop()!,
  result: fixture.result.structuredContent as FlintChartResult,
}));

const catalog = Object.values(
  import.meta.glob<{ chart_types: { name: string; channels: string[] }[] }>(
    "../../../runtime/ai/chartquery/flint_catalog.json",
    { eager: true, import: "default" },
  ),
)[0];

function registerFormatters(chart: AdaptedFlintChart) {
  expressionFunction("humanize", (value: number) =>
    humanizeDataType(value, FormatPreset.HUMANIZE, "table"),
  );
  for (const measure of chart.measures) {
    const formatter = createMeasureValueFormatter<null | undefined>(measure);
    expressionFunction(
      sanitizeFieldName(measure.name ?? "measure"),
      (value: number) => formatter(value) ?? "",
    );
  }
}

async function runHeadless(chart: AdaptedFlintChart, isDark = false) {
  registerFormatters(chart);
  const compiled = compileFlintSpec({
    spec: chart.flint.spec,
    rows: chart.rows,
    fields: chart.flint.fields,
    size: { width: 600, height: 320 },
    measureFields: chart.flint.measureFields,
  });
  expect(compiled.error).toBeNull();
  expect(compiled.spec).not.toBeNull();
  const { spec: vgSpec } = compile(compiled.spec as TopLevelSpec, {
    config: getRillTheme(isDark, chart.theme),
  });
  const runtime = parse(vgSpec, undefined, { ast: true });
  const view = new View(runtime, {
    renderer: "none",
    expr: expressionInterpreter,
  });
  await view.runAsync();
  return { view, compiled };
}

describe("Flint render path", () => {
  it("covers every Flint fixture", () => {
    expect(flintFixtures.length).toBeGreaterThanOrEqual(8);
  });

  it("uses chart types and channels the server's catalog knows", () => {
    expect(catalog?.chart_types.length).toBeGreaterThan(20);
    for (const { file, result } of flintFixtures) {
      const entry = catalog.chart_types.find(
        (type) => type.name === result.input.chart_spec.chartType,
      );
      expect(entry, file).toBeDefined();
      for (const channel of Object.keys(result.input.chart_spec.encodings)) {
        expect(entry!.channels, `${file}: ${channel}`).toContain(channel);
      }
    }
  });

  for (const { file, result } of flintFixtures) {
    it(`compiles and runs the ${file} fixture`, async () => {
      const chart = adaptFlintResult(result, { isDark: false, locale: "en" });
      if (chart.rowCount === 0) return;
      const { view, compiled } = await runHeadless(chart);
      expect(compiled.warnings.filter((w) => w.severity === "error")).toEqual(
        [],
      );
      expect(view.width()).toBeGreaterThan(0);
      expect(view.height()).toBeGreaterThan(0);
    });
  }

  it("draws the comparison as a second series with the previous-period label", async () => {
    const chart = adaptFlintResult(
      comparison.result.structuredContent as FlintChartResult,
      { isDark: false, locale: "en" },
    );
    expect(chart.hasComparison).toBe(true);
    expect(chart.flint.comparison).toEqual({
      channel: "color",
      measure: "total_bids",
    });
    const { compiled } = await runHeadless(chart);
    const spec = compiled.spec as unknown as {
      encoding?: { color?: { field?: string; title?: string } };
      data?: { values?: Record<string, unknown>[] };
    };
    expect(spec.encoding?.color?.field).toBe("rill_period");
    expect(spec.encoding?.color?.title).toBe("Period");
    const periods = new Set(
      (spec.data?.values ?? []).map((row) => row.rill_period),
    );
    expect(periods).toEqual(
      new Set(["Total Bids", "Total Bids (previous period)"]),
    );
    // The table keeps the result's own rows while the chart draws the folded ones.
    expect(chart.table.totalRows).toBe(chart.rows.length / 2);
  });

  it("routes measure axes through Rill's formatters", async () => {
    const chart = adaptFlintResult(
      comparison.result.structuredContent as FlintChartResult,
      { isDark: true, locale: "en" },
    );
    const { compiled } = await runHeadless(chart, true);
    expect(JSON.stringify(compiled.spec)).toContain(
      `"formatType":"${sanitizeFieldName("total_bids")}"`,
    );
  });
});
