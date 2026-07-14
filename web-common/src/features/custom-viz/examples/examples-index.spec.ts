import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";
import {
  exampleComponentYAML,
  loadGallery,
  type GalleryExample,
} from "./example-to-component";

// Integrity checks for the vendored Vega-Lite examples snapshot: every entry must
// render self-contained (offline), so no external data references may remain.

describe("vega-examples snapshot", async () => {
  const gallery = await loadGallery();

  it("has a usable number of examples and a license notice", () => {
    expect(gallery.examples.length).toBeGreaterThan(30);
    expect(gallery.license).toContain("BSD-3-Clause");
  });

  it("has unique example names (they key UI lists and generated file names)", () => {
    const names = gallery.examples.map((example) => example.name);
    expect(new Set(names).size).toBe(names.length);
  });

  it("every example parses, has inline data, and carries gallery metadata", () => {
    for (const example of gallery.examples) {
      expect(example.name, "missing name").toBeTruthy();
      expect(example.title, `missing title for ${example.name}`).toBeTruthy();
      expect(
        example.subcategory,
        `missing subcategory for ${example.name}`,
      ).toBeTruthy();

      const spec = JSON.parse(example.spec);
      expect(spec, `spec is not an object for ${example.name}`).toBeTypeOf(
        "object",
      );
      // No remaining external references anywhere in the spec.
      expect(
        example.spec.includes('"url"'),
        `external data url remains in ${example.name}`,
      ).toBe(false);
    }
  });

  it("converts an example into valid component YAML", () => {
    const example = gallery.examples[0] as GalleryExample;
    const yaml = exampleComponentYAML(example);
    const doc = parseDocument(yaml);
    expect(doc.errors).toEqual([]);

    const parsed = doc.toJSON();
    expect(parsed.type).toBe("component");
    expect(parsed.display_name).toBe(example.title);
    expect(() => JSON.parse(parsed.custom_chart.vega_spec)).not.toThrow();
  });

  it("parameterizes the simple bar chart with role-named typed params and templated SQL", () => {
    const bar = gallery.examples.find((example) => example.name === "bar");
    expect(bar).toBeDefined();

    const parsed = parseDocument(exampleComponentYAML(bar!)).toJSON();

    // Typed params named after the chart role (not the example's column names);
    // field params are required with no defaults, plus a customizable sort field
    // and a row limit whose default matches the chart's density (bars → 10).
    expect(parsed.params.map((p: Record<string, unknown>) => p.name)).toEqual([
      "metrics_view",
      "x_axis",
      "y_axis",
      "order_by",
      "limit",
    ]);
    expect(parsed.params.map((p: Record<string, unknown>) => p.type)).toEqual([
      "metrics_view",
      "dimension",
      "measure",
      "string",
      "number",
    ]);
    for (const param of parsed.params.slice(0, 4)) {
      expect(param.required).toBe(true);
      expect(param.default).toBeUndefined();
    }
    expect(parsed.params[4].default).toBe(10);
    expect(parsed.params[4].required).toBeUndefined();
    expect(parsed.params[1].description).toContain('"a"');

    // The query is fully templated over the params, on a single line, and is a
    // Top-N over the order_by param (time series order by time instead).
    expect(parsed.custom_chart.metrics_sql).toEqual([
      "SELECT {{ .params.x_axis }}, {{ .params.y_axis }} FROM {{ .params.metrics_view }} ORDER BY {{ .params.order_by }} DESC LIMIT {{ .params.limit }}",
    ]);

    // The spec keeps only the visual design: sample data is gone, data is bound
    // to the query, and field references are templated.
    const spec = JSON.parse(parsed.custom_chart.vega_spec);
    expect(spec.data).toEqual({ name: "query1" });
    expect(JSON.stringify(spec)).not.toContain('"values"');
    expect(spec.encoding.x.field).toBe("{{ .params.x_axis }}");
    expect(spec.encoding.y.field).toBe("{{ .params.y_axis }}");
  });

  it("sizes the default row limit to the chart's visual density", () => {
    // Scatterplots draw one small mark per row and stay readable with many
    // rows (default 100); bar charts default to 10 (asserted in the bar test above).
    const scatter = gallery.examples.find(
      (example) => example.name === "point_2d",
    );
    if (!scatter) return; // Snapshot refresh may drop it; nothing to assert then.
    const parsed = parseDocument(exampleComponentYAML(scatter)).toJSON();
    const limit = parsed.params.find(
      (p: Record<string, unknown>) => p.name === "limit",
    );
    expect(limit).toMatchObject({ type: "number", default: 100 });
    expect(parsed.custom_chart.metrics_sql[0]).toContain(
      "LIMIT {{ .params.limit }}",
    );
  });

  it("collapses timeUnit channels over one date field into a single time_dimension param", () => {
    // The weather heatmap encodes x (day of month) and y (month) from the SAME
    // date field via timeUnits; both channels must share one time_dimension param
    // with their timeUnits preserved, and the query orders by time ascending.
    const weather = gallery.examples.find(
      (example) => example.name === "rect_heatmap_weather",
    );
    if (!weather) return; // Snapshot refresh may drop it; nothing to assert then.
    const parsed = parseDocument(exampleComponentYAML(weather)).toJSON();

    const timeParams = parsed.params.filter(
      (p: Record<string, unknown>) => p.type === "time_dimension",
    );
    expect(timeParams).toHaveLength(1);
    const timeRef = `{{ .params.${timeParams[0].name} }}`;

    const spec = JSON.parse(parsed.custom_chart.vega_spec);
    expect(spec.encoding.x.field).toBe(timeRef);
    expect(spec.encoding.y.field).toBe(timeRef);
    expect(spec.encoding.x.timeUnit).toBe("date");
    expect(spec.encoding.y.timeUnit).toBe("month");
    expect(spec.encoding.color.type).toBe("quantitative");

    expect(parsed.custom_chart.metrics_sql[0]).toContain(
      `ORDER BY ${timeRef} LIMIT {{ .params.limit }}`,
    );
    const limit = parsed.params.find(
      (p: Record<string, unknown>) => p.name === "limit",
    );
    expect(limit).toMatchObject({ type: "number", default: 2000 });
  });

  it("falls back to a verbatim sample import for value-dependent specs", () => {
    // The diverging stacked bar example bakes the sample data's category values
    // into calculate/stack transforms; it cannot be remapped onto arbitrary data.
    const diverging = gallery.examples.find((example) =>
      example.name.includes("bar_diverging"),
    );
    if (!diverging) return; // Snapshot refresh may drop it; nothing to assert then.
    const parsed = parseDocument(exampleComponentYAML(diverging)).toJSON();
    expect(parsed.params).toBeUndefined();
    expect(parsed.custom_chart.metrics_sql).toBeUndefined();
    expect(parsed.custom_chart.vega_spec).toContain('"values"');
  });

  it("parameterizes a healthy share of the gallery deterministically", () => {
    let converted = 0;
    for (const example of gallery.examples) {
      const parsed = parseDocument(exampleComponentYAML(example)).toJSON();
      if (!parsed.params) continue; // Fallback import (sample data kept).
      converted++;

      // Converted components must be fully templated and data-free.
      expect(parsed.params[0]).toEqual({
        name: "metrics_view",
        type: "metrics_view",
        required: true,
      });
      for (const param of parsed.params) {
        expect([
          "metrics_view",
          "dimension",
          "measure",
          "time_dimension",
          "string",
          "number",
        ]).toContain(param.type);
        // Field params and order_by never carry defaults; the limit param must.
        if (param.type === "number") {
          expect(param.default, `param ${param.name} of ${example.name}`)
            .toBeTypeOf("number");
        } else {
          expect(param.default, `param ${param.name} of ${example.name}`)
            .toBeUndefined();
        }
      }
      expect(parsed.custom_chart.metrics_sql[0]).toContain(
        "FROM {{ .params.metrics_view }}",
      );
      const spec = JSON.parse(parsed.custom_chart.vega_spec);
      expect(JSON.stringify(spec.data)).toBe('{"name":"query1"}');
    }
    // Transform-based examples fall back, so not everything converts;
    // a meaningful floor keeps the deterministic path honest.
    expect(converted).toBeGreaterThan(20);
  });

  it("normalizes single-view sizes to fill the container", () => {
    const single = gallery.examples.find((example) => {
      const spec = JSON.parse(example.spec);
      return !["hconcat", "vconcat", "concat", "facet", "repeat", "spec"].some(
        (key) => key in spec,
      );
    });
    expect(single).toBeDefined();
    const yaml = exampleComponentYAML(single!);
    const spec = JSON.parse(
      parseDocument(yaml).toJSON().custom_chart.vega_spec,
    );
    expect([spec.width, typeof spec.width]).toSatisfy(
      ([width, type]: [unknown, string]) =>
        width === "container" || type === "object",
    );
  });
});
