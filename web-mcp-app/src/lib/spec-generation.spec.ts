import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import { generateSpec } from "@rilldata/web-common/features/components/charts/spec-generators";
import {
  createMeasureValueFormatter,
  humanizeDataType,
} from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { expressionFunction, parse, View } from "vega";
import { expressionInterpreter } from "vega-interpreter";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import { adaptChartResult, type AdaptedRillChart } from "./adapt";
import type { RillChartResult } from "./contract";

// The widget registers one Vega expression function per measure (axis and tooltip
// formatting) plus `humanize`; mirror that so the compiled specs resolve them.
function registerFormatters(chart: AdaptedRillChart) {
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

/** Compiles the Vega-Lite spec and evaluates it with the CSP-safe interpreter, no renderer. */
async function runHeadless(chart: AdaptedRillChart, isDark: boolean) {
  registerFormatters(chart);
  const vlSpec = generateSpec(
    chart.chartType,
    chart.spec,
    chart.snapshot,
  ) as TopLevelSpec;
  const { spec: vgSpec } = compile(
    { ...vlSpec, width: 600, height: 320 } as TopLevelSpec,
    { config: getRillTheme(isDark, chart.theme) },
  );
  const runtime = parse(vgSpec, undefined, { ast: true });
  const view = new View(runtime, {
    renderer: "none",
    expr: expressionInterpreter,
  });
  view.data("metrics-view", chart.snapshot.data);
  await view.runAsync();
  return view;
}

interface Fixture {
  name: string;
  expect: string;
  result: { structuredContent?: RillChartResult };
}

const fixtures = Object.entries(
  import.meta.glob<Fixture>("../../harness/fixtures/*.json", {
    eager: true,
    import: "default",
  }),
)
  .map(([path, fixture]) => ({ file: path.split("/").pop()!, fixture }))
  // Flint results (`input`) take the Flint render path; see flint-compile.spec.ts.
  .filter(
    ({ fixture }) =>
      fixture.expect === "canvas" && !fixture.result.structuredContent?.input,
  );

describe("Vega-Lite spec generation from fixtures", () => {
  it("covers every renderable fixture", () => {
    expect(fixtures.length).toBeGreaterThanOrEqual(13);
  });

  for (const mode of [false, true]) {
    for (const { file, fixture } of fixtures) {
      it(`${file} produces a serializable spec (${mode ? "dark" : "light"})`, () => {
        const chart = adaptChartResult(fixture.result.structuredContent!, {
          isDark: mode,
          locale: "en",
        });
        const spec = generateSpec(chart.chartType, chart.spec, chart.snapshot);
        expect(spec).toBeTypeOf("object");
        const serialized = JSON.stringify(spec);
        expect(serialized.length).toBeGreaterThan(50);
        // Specs reference the data by name, never inline it.
        expect(serialized).toContain('"metrics-view"');
      });
    }
  }

  for (const { file, fixture } of fixtures) {
    it(`${file} compiles and runs through the Vega interpreter`, async () => {
      const chart = adaptChartResult(fixture.result.structuredContent!, {
        isDark: false,
        locale: "en",
      });
      const view = await runHeadless(chart, false);
      expect(view.data("metrics-view")).toHaveLength(chart.rowCount);
    });
  }

  it("keeps the backend's category order in the bar spec", () => {
    const { fixture } = fixtures.find(({ file }) => file === "bar_chart.json")!;
    const chart = adaptChartResult(fixture.result.structuredContent!, {
      isDark: false,
    });
    const serialized = JSON.stringify(
      generateSpec(chart.chartType, chart.spec, chart.snapshot),
    );
    expect(serialized).toContain('"sort":["Acme Media","Beta Brands"');
  });
});
