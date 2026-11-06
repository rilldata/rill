import { compileFlintSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/bar_chart.json";
import grouped from "../../harness/fixtures/grouped_bar.json";
import line from "../../harness/fixtures/line_chart.json";
import comparison from "../../harness/fixtures/line_comparison.json";
import pie from "../../harness/fixtures/pie.json";
import { adaptResult } from "./adapt";
import { normalizeRows, type ChartResult } from "./contract";
import { buildFlintCompileInput } from "./flint-input";
import {
  buildFlintPanel,
  layoutLabel,
  MAX_PANEL_PROPERTIES,
  valueKey,
  withProperty,
} from "./flint-options";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as ChartResult;

function panelFor(fixture: { result: { structuredContent?: unknown } }) {
  const res = result(fixture);
  const rows = normalizeRows(res.data);
  const compile = buildFlintCompileInput(
    res.input.chart_spec,
    res,
    rows,
    undefined,
  );
  return {
    res,
    rows,
    compile,
    panel: buildFlintPanel(res.input.chart_spec, compile, res.fields ?? {}),
  };
}

const choicesOf = (panel: ReturnType<typeof buildFlintPanel>) =>
  panel.chartType?.choices.map((choice) => choice.value);

describe("buildFlintPanel", () => {
  it("offers Flint's chart-type transitions, the current type first", () => {
    const { res, rows, panel } = panelFor(bar);
    // A bar chart over an unordered dimension only moves to a lollipop: a line needs an ordered axis.
    expect(choicesOf(panel)).toEqual(["Bar Chart", "Lollipop Chart"]);
    expect(choicesOf(panelFor(line).panel)).toEqual([
      "Line Chart",
      "Area Chart",
      "Bar Chart",
    ]);
    for (const type of choicesOf(panel)!) {
      const moved = panel.apply.chartType(type);
      const movedCompile = buildFlintCompileInput(moved, res, rows, undefined);
      expect(
        compileFlintSpec({
          spec: movedCompile.spec,
          rows,
          fields: movedCompile.fields,
          size: { width: 600, height: 320 },
          measureFields: movedCompile.measureFields,
        }).error,
        type,
      ).toBeNull();
    }

    const moved = panel.apply.chartType("Lollipop Chart");
    expect(moved.chartType).toBe("Lollipop Chart");
    // The top-N limit follows the field; the sort keys of the old type do not.
    expect(moved.encodings).toEqual({
      x: { field: "advertiser_name", limit: 10 },
      y: "total_bids",
    });
    expect(moved).not.toHaveProperty("chartProperties");
    // The result's spec is untouched, and choosing it again restores it.
    expect(res.input.chart_spec.chartType).toBe("Bar Chart");
    expect(panel.apply.chartType("Bar Chart")).toBe(res.input.chart_spec);
  });

  it("counts the categories of a time axis by time rather than by row", () => {
    // Two series over 20 days: 40 rows, but 20 buckets, few enough for bars. The adapter hands
    // Flint Date objects, which a distinct count would otherwise tell apart by identity.
    const days = Array.from(
      { length: 20 },
      (_, i) => `2023-09-${String(i + 1).padStart(2, "0")}T00:00:00Z`,
    );
    const range = { start: days[0], end: "2023-09-21T00:00:00Z" };
    const fields = {
      __time: { name: "__time", kind: "time" as const },
      device: { name: "device", kind: "dimension" as const },
      bids: { name: "bids", kind: "measure" as const },
    };
    const chart = adaptResult(
      {
        widget: "chart",
        chart_type: "Line Chart",
        input: {
          chart_spec: {
            chartType: "Line Chart",
            encodings: { x: "__time", y: "bids", color: "device" },
          },
          data: { metrics_view: "bids", time_range: range },
        },
        data: days.flatMap((__time, i) =>
          ["Mobile", "Desktop"].map((device) => ({ __time, device, bids: i })),
        ),
        time_grain: "TIME_GRAIN_DAY",
        fields,
        resolved_time_range: range,
        metrics_view: { name: "bids", time_dimension: "__time" },
      },
      { isDark: false },
    );
    expect(chart.rows[0].__time).toBeInstanceOf(Date);
    expect(
      choicesOf(buildFlintPanel(chart.chartSpec, chart.flint, fields)),
    ).toContain("Bar Chart");
  });

  it("re-routes the fields the way Flint's transitions do", () => {
    const { panel } = panelFor(grouped);
    expect(choicesOf(panel)).toContain("Stacked Bar Chart");
    // The clusters on group become the stacks on color, and keep their limit.
    expect(panel.apply.chartType("Stacked Bar Chart").encodings).toEqual({
      x: { field: "publisher", limit: 10 },
      y: "impressions",
      color: { field: "device", limit: 10 },
    });
  });

  it("offers only the types that draw the comparison while one is drawn", () => {
    const { res, rows, compile, panel } = panelFor(comparison);
    expect(compile.comparison).toBeDefined();
    const choices = choicesOf(panel)!;
    expect(choices[0]).toBe("Line Chart");
    expect(choices.length).toBeGreaterThan(1);
    for (const type of choices) {
      const moved = buildFlintCompileInput(
        panel.apply.chartType(type),
        res,
        rows,
        undefined,
      );
      expect(moved.comparison, type).toBeDefined();
    }
    // A bar chart draws the periods side by side.
    const bars = buildFlintCompileInput(
      panel.apply.chartType("Bar Chart"),
      res,
      rows,
      undefined,
    );
    expect(bars.spec.chartType).toBe("Grouped Bar Chart");
  });

  it("keeps the choices of the result's spec while another type is shown", () => {
    const { res, rows, panel } = panelFor(bar);
    const alternative = panel.chartType!.choices[1].value;
    const moved = panel.apply.chartType(alternative);
    const movedCompile = buildFlintCompileInput(moved, res, rows, undefined);
    const next = buildFlintPanel(
      moved,
      movedCompile,
      res.fields ?? {},
      res.input.chart_spec,
    );
    expect(next.chartType!.current).toBe(alternative);
    expect(next.chartType!.choices.map((choice) => choice.value)).toEqual(
      panel.chartType!.choices.map((choice) => choice.value),
    );
  });

  it("keeps the applicable binary and discrete properties only", () => {
    const { panel } = panelFor(pie);
    expect(panel.properties.length).toBeLessThanOrEqual(MAX_PANEL_PROPERTIES);
    for (const property of panel.properties) {
      expect(property.choices.length).toBeGreaterThan(1);
      expect(property.key).not.toBe("innerRadius"); // continuous
    }
    const sortSlices = panel.properties.find((p) => p.key === "sortSlices");
    expect(sortSlices).toBeDefined();
    const next = panel.apply.property(
      "sortSlices",
      sortSlices!.choices[1].value,
    );
    expect(next.chartProperties).toEqual({
      innerRadius: 55,
      sortSlices: JSON.parse(sortSlices!.choices[1].value),
    });
  });

  it("offers the arrangements and writes the chosen one into the properties", () => {
    const { panel } = panelFor(bar);
    expect(panel.layout).toBeDefined();
    expect(panel.layout!.choices[0].value).toBe(panel.layout!.value);
    expect(panel.layout!.choices.map((choice) => choice.label)).toEqual([
      "Default",
      "Swap axes",
    ]);
    const flipped = panel.apply.layout(panel.layout!.choices[1].value);
    expect(flipped.chartProperties).toEqual({
      [panel.layout!.key]: panel.layout!.choices[1].value,
    });
    // Choosing the default view again clears the override.
    expect(
      panel.apply.layout(panel.layout!.choices[0].value).chartProperties,
    ).toEqual({});
  });

  it("leaves chart-type moves to the chart-type control, and the compiler honors every arrangement", () => {
    for (const fixture of [bar, grouped, comparison]) {
      const { res, rows, panel } = panelFor(fixture);
      const compiled = (spec: typeof res.input.chart_spec) => {
        const input = buildFlintCompileInput(spec, res, rows, undefined);
        return JSON.stringify(
          compileFlintSpec({
            spec: input.spec,
            rows: input.rows,
            fields: input.fields,
            size: { width: 600, height: 320 },
            measureFields: input.measureFields,
          }).spec,
        );
      };
      const authored = compiled(res.input.chart_spec);
      for (const choice of panel.layout?.choices.slice(1) ?? []) {
        expect(choice.value).not.toContain("type:");
        expect(choice.label).not.toMatch(/θ|⇄/);
        expect(
          compiled(panel.apply.layout(choice.value)),
          choice.value,
        ).not.toBe(authored);
      }
    }
  });

  it("shows the default choice of a discrete property as selected", () => {
    const { panel } = panelFor(comparison);
    const curve = panel.properties.find((p) => p.key === "interpolate");
    expect(curve).toBeDefined();
    expect(curve!.value).toBe(curve!.choices[0].value);
    const smooth = panel.apply.property("interpolate", '"monotone"');
    expect(smooth.chartProperties).toEqual({ interpolate: "monotone" });
    // Choosing the default again clears the property instead of storing null.
    expect(
      panel.apply.property("interpolate", curve!.choices[0].value)
        .chartProperties,
    ).toEqual({});
    for (const property of panel.properties) {
      expect(
        property.choices.map((choice) => choice.value),
        property.key,
      ).toContain(property.value);
    }
  });
});

describe("layoutLabel", () => {
  it("labels each step of a state and keeps Flint's label for an unknown one", () => {
    expect(layoutLabel("default", "Default")).toBe("Default");
    expect(layoutLabel("flip:x-y", "X ⇄ Y")).toBe("Swap axes");
    expect(layoutLabel("swap:x-color", "X ⇄ Color")).toBe(
      "Swap x-axis and color",
    );
    expect(layoutLabel("flip:x-y|augment:row", "X ⇄ Y · Color + Rows")).toBe(
      "Swap axes · Split into rows",
    );
    expect(layoutLabel("series:color", "Columns ⇄ Color")).toBe(
      "Combine panels",
    );
    expect(layoutLabel("flip:x-y|twist:z", "X ⇄ Y · Twist")).toBe(
      "X ⇄ Y · Twist",
    );
  });
});

describe("withProperty and valueKey", () => {
  it("sets, replaces and clears a property", () => {
    const spec = { chartType: "Bar Chart", encodings: { x: "a", y: "b" } };
    const set = withProperty(spec, "cornerRadius", 4);
    expect(set.chartProperties).toEqual({ cornerRadius: 4 });
    expect(
      withProperty(set, "cornerRadius", undefined).chartProperties,
    ).toEqual({});
    expect(spec).not.toHaveProperty("chartProperties");
    expect(valueKey(undefined)).toBe("null");
    expect(valueKey("stacked")).toBe('"stacked"');
  });
});
