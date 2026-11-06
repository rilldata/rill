import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import { compileFlintSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
import {
  createMeasureValueFormatter,
  humanizeDataType,
} from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import { dirname, resolve } from "node:path";
import { expressionFunction, parse, View } from "vega";
import { expressionInterpreter } from "vega-interpreter";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import comparison from "../../harness/fixtures/flint_line_comparison.json";
import { adaptFlintResult, type AdaptedFlintChart } from "./adapt";
import type { FlintChartResult, WireRow } from "./contract";

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
  import.meta.glob<{
    flint_version: string;
    chart_types: { name: string; channels: string[] }[];
  }>("../../../runtime/ai/chartquery/flint_catalog.json", {
    eager: true,
    import: "default",
  }),
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
  return { view, compiled, vgSpec };
}

describe("Flint render path", () => {
  it("covers every Flint fixture", () => {
    expect(flintFixtures.length).toBeGreaterThanOrEqual(8);
  });

  it("validates with a catalog generated from the installed Flint", () => {
    // The package does not export its package.json, so it is read next to the resolved entry
    // point, as scripts/generate-flint-catalog.mjs does.
    const require = createRequire(import.meta.url);
    const packageRoot = resolve(
      dirname(require.resolve("flint-chart/vegalite")),
      "../..",
    );
    const { version } = JSON.parse(
      readFileSync(resolve(packageRoot, "package.json"), "utf8"),
    ) as { version: string };
    expect(
      catalog.flint_version,
      "regenerate the catalog: npm run generate:catalog -w web-mcp-app",
    ).toBe(version);
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

  it("cumulates an ECDF over the rows rather than per detail value", async () => {
    // The server's `detail` is the row grain: one row per app.
    const data: WireRow[] = [0.1, 0.4, 0.2, 0.8, 0.5, 0.3].map((ctr, i) => ({
      app: `app ${i}`,
      platform: i % 2 === 0 ? "App" : "Site",
      ctr,
    }));
    const ecdf = (encodings: Record<string, string>) =>
      adaptFlintResult(
        {
          widget: "chart",
          chart_type: "ECDF Plot",
          input: {
            chart_spec: { chartType: "ECDF Plot", encodings },
            data: {
              metrics_view: "bids",
              time_range: {
                start: "2023-09-09T00:00:00Z",
                end: "2023-09-16T00:00:00Z",
              },
            },
          },
          data,
          fields: {
            app: { name: "app", kind: "dimension" },
            platform: { name: "platform", kind: "dimension" },
            ctr: { name: "ctr", kind: "measure", format_preset: "percentage" },
          },
        },
        { isDark: false, locale: "en" },
      );
    const proportions = async (chart: AdaptedFlintChart) => {
      const { view, vgSpec } = await runHeadless(chart);
      return (vgSpec.data ?? [])
        .flatMap((dataset) => view.data(dataset.name) as { __ecdf?: number }[])
        .map((row) => row.__ecdf)
        .filter((value): value is number => typeof value === "number");
    };

    const single = await proportions(ecdf({ x: "ctr", detail: "app" }));
    expect(new Set(single)).toEqual(
      new Set([1, 2, 3, 4, 5, 6].map((n) => n / 6)),
    );
    // A color split still draws one curve per value, each over its own rows.
    const split = await proportions(
      ecdf({ x: "ctr", detail: "app", color: "platform" }),
    );
    expect(new Set(split)).toEqual(new Set([1 / 3, 2 / 3, 1]));
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
