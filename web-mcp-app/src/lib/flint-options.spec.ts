import { compileFlintSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/flint_bar_chart.json";
import pie from "../../harness/fixtures/flint_pie.json";
import { normalizeRows, type FlintChartResult } from "./contract";
import { buildFlintCompileInput } from "./flint-input";
import {
  buildFlintPanel,
  EXCLUDED_CHART_TYPES,
  MAX_CHART_TYPE_CANDIDATES,
  MAX_PANEL_PROPERTIES,
  valueKey,
  withChartType,
  withProperty,
} from "./flint-options";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as FlintChartResult;

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
    panel: buildFlintPanel(res.input.chart_spec, compile, rows),
  };
}

describe("buildFlintPanel", () => {
  it("offers chart types the bindings can move to, the current one first", () => {
    const { res, rows, panel } = panelFor(bar);
    expect(panel.chartType).toBeDefined();
    const choices = panel.chartType!.choices.map((choice) => choice.value);
    expect(choices[0]).toBe("Bar Chart");
    expect(choices.length).toBeGreaterThan(1);
    expect(choices.length).toBeLessThanOrEqual(MAX_CHART_TYPE_CANDIDATES + 1);
    expect(new Set(choices).size).toBe(choices.length);
    for (const type of choices) {
      expect(EXCLUDED_CHART_TYPES.has(type)).toBe(false);
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

    const moved = panel.apply.chartType(choices[1]);
    expect(moved.chartType).toBe(choices[1]);
    expect(Object.keys(moved.encodings).length).toBeGreaterThan(0);
    // The result's spec is untouched, and choosing it again restores it.
    expect(res.input.chart_spec.chartType).toBe("Bar Chart");
    expect(panel.apply.chartType("Bar Chart")).toBe(res.input.chart_spec);
  });

  it("keeps the choices of the result's spec while another type is shown", () => {
    const { res, rows, panel } = panelFor(bar);
    const alternative = panel.chartType!.choices[1].value;
    const moved = panel.apply.chartType(alternative);
    const movedCompile = buildFlintCompileInput(moved, res, rows, undefined);
    const next = buildFlintPanel(
      moved,
      movedCompile,
      rows,
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

  it("offers the pivot views and writes the chosen state into the properties", () => {
    const { panel } = panelFor(bar);
    expect(panel.layout).toBeDefined();
    expect(panel.layout!.choices[0].value).toBe(panel.layout!.value);
    const flipped = panel.apply.layout(panel.layout!.choices[1].value);
    expect(flipped.chartProperties).toEqual({
      pivot: panel.layout!.choices[1].value,
    });
    // Choosing the default view again clears the override.
    expect(
      panel.apply.layout(panel.layout!.choices[0].value).chartProperties,
    ).toEqual({});
  });
});

describe("withChartType", () => {
  it("re-binds the fields and keeps the limit of a moved field", () => {
    const { res, rows, compile } = panelFor(bar);
    const moved = withChartType(
      res.input.chart_spec,
      "Lollipop Chart",
      rows,
      compile.fields.semantic_types,
    );
    expect(moved.chartType).toBe("Lollipop Chart");
    // The top-N limit follows the field; the sort keys of the old type do not.
    expect(moved.encodings.x).toEqual({ field: "advertiser_name", limit: 10 });
    expect(moved.encodings.y).toBe("total_bids");
    expect(moved).not.toHaveProperty("chartProperties");
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
