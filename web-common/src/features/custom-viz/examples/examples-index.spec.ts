import { describe, expect, it } from "vitest";
import { parse } from "yaml";
import { vlGetTemplateDef } from "flint-chart/vegalite";
import index from "./flint-charts-index.json";
import { exampleComponentYAML, type GalleryEntry } from "./example-spec";

/**
 * Guards the generated chart catalog (scripts/generate-flint-catalog.ts) against drift from the
 * installed flint-chart, and the entry → component YAML conversion against the shape the Go parser
 * and the renderer expect.
 */

const charts = index.examples as GalleryEntry[];

function chartNamed(chartType: string): GalleryEntry {
  const entry = charts.find((chart) => chart.chartType === chartType);
  if (!entry) throw new Error(`${chartType} missing from the catalog`);
  return entry;
}

const chartTypes = new Set(charts.map((chart) => chart.chartType));

function componentFor(entry: GalleryEntry) {
  return parse(exampleComponentYAML(entry)) as {
    type: string;
    display_name: string;
    params: {
      name: string;
      type: string;
      required?: boolean;
      default?: unknown;
    }[];
    custom_chart: {
      metrics_sql: string;
      spec: { chartType: string; encodings: Record<string, { field: string }> };
    };
  };
}

function paramRefs(text: string): string[] {
  return [...text.matchAll(/\{\{\s*\.params\.(\w+)\s*\}\}/g)].map((m) => m[1]);
}

describe("flint chart catalog", () => {
  it("covers the chart types Flint can render, a few examples each", () => {
    expect(chartTypes.size).toBeGreaterThan(25);
    expect(index.flintVersion).toMatch(/^\d+\.\d+\.\d+/);
  });

  it("keeps the per-chart-type count small enough to browse", () => {
    const counts = new Map<string, number>();
    for (const chart of charts) {
      counts.set(chart.chartType, (counts.get(chart.chartType) ?? 0) + 1);
    }
    for (const [chartType, count] of counts) {
      // No chart type should flood the grid. A single card is allowed: Flint's other examples
      // for a type can need fields a metrics view never has (a Gantt chart's two time dimensions).
      expect(count, chartType).toBeLessThanOrEqual(3);
    }
  });

  it("excludes Flint's adversarial rendering tests", () => {
    // These examples exist to overflow an axis, exercise a degenerate shape, or assert a warning
    // fires. Offering one as a starter would hand the user a chart that looks broken.
    const adversarial = [
      "stress",
      "overflow",
      "cutoff",
      "edge-case",
      "degenerate",
      "warning",
      "dtype-conversion",
    ];
    for (const chart of charts) {
      for (const tag of adversarial) {
        expect(chart.tags, `${chart.name} is tagged ${tag}`).not.toContain(tag);
      }
    }
  });

  it("excludes the chart types Rill does not offer", () => {
    // Maps need semantics Rill never derives, and a KPI card duplicates the canvas widget.
    for (const chartType of ["Map", "Choropleth", "KPI Card"]) {
      expect(chartTypes, chartType).not.toContain(chartType);
    }
  });

  it("gives every example a title and a description to browse by", () => {
    for (const chart of charts) {
      expect(chart.title, chart.name).toBeTruthy();
      // The card leads with the chart type and reads the description underneath, so an example
      // without one gives the reader nothing to tell it apart from its siblings.
      expect(chart.description, chart.name).toBeTruthy();
    }
  });

  it("has unique names, which key UI lists and generated file names", () => {
    expect(new Set(charts.map((chart) => chart.name)).size).toBe(charts.length);
  });

  it("only lists chart types the installed flint-chart can render", () => {
    for (const chart of charts) {
      expect(
        () => vlGetTemplateDef(chart.chartType),
        `${chart.chartType} is not a registered Flint chart type`,
      ).not.toThrow();
    }
  });

  it("only binds channels the chart type declares", () => {
    for (const chart of charts) {
      const declared = vlGetTemplateDef(chart.chartType) as {
        channels: string[];
      };
      for (const binding of chart.bindings) {
        expect(
          declared.channels,
          `${chart.chartType} binds ${binding.channel}, which it does not accept`,
        ).toContain(binding.channel);
      }
    }
  });

  it("inlines every thumbnail as a self-contained data URI", () => {
    for (const chart of charts) {
      // web-common is aliased in from outside each app's Vite root, so an asset URL resolved
      // against web-common 404s in web-local and web-admin. The SVG has to travel in the module.
      expect(chart.thumbnail, chart.name).toMatch(/^data:image\/svg\+xml,/);
      expect(chart.thumbnail, chart.name).toContain("%3Csvg");
      // A raw quote or angle bracket would terminate the src attribute.
      expect(chart.thumbnail, chart.name).not.toMatch(/["<>]/);
    }
  });

  it("binds every chart to at least one field, with a valid param role", () => {
    const roles = ["measure", "dimension", "time_dimension"];
    for (const chart of charts) {
      expect(chart.bindings.length, chart.chartType).toBeGreaterThan(0);
      for (const binding of chart.bindings) {
        expect(roles, `${chart.chartType}.${binding.channel}`).toContain(
          binding.role,
        );
      }
    }
  });

  it("spreads charts across several categories", () => {
    expect(new Set(charts.map((chart) => chart.category)).size).toBeGreaterThan(
      1,
    );
  });
});

describe("exampleComponentYAML", () => {
  it("produces the component shape the parser requires", () => {
    const component = componentFor(chartNamed("Bar Chart"));

    expect(component.type).toBe("component");
    // metrics_sql and spec must be inside custom_chart, never at the top level.
    expect(component.custom_chart.metrics_sql).toBeTruthy();
    expect(component.custom_chart.spec.chartType).toBe("Bar Chart");
    expect(component).not.toHaveProperty("metrics_sql");
    expect(component).not.toHaveProperty("spec");
    // Vega-Lite is not authored at the component level.
    expect(component.custom_chart).not.toHaveProperty("vega_spec");
  });

  it("declares a metrics_view param plus one typed param per bound channel", () => {
    const entry = chartNamed("Bar Chart");
    const component = componentFor(entry);
    const byName = new Map(component.params.map((p) => [p.name, p]));

    expect(byName.get("metrics_view")).toMatchObject({
      type: "metrics_view",
      required: true,
    });

    for (const binding of entry.bindings) {
      expect(byName.get(binding.param), binding.param).toMatchObject({
        type: binding.role,
        required: true,
      });
    }

    expect(byName.get("limit")).toMatchObject({
      type: "number",
      default: entry.defaultLimit,
    });
  });

  it("templates every encoding against a declared param", () => {
    for (const entry of charts) {
      const component = componentFor(entry);
      const declared = component.params.map((p) => p.name);

      for (const [channel, encoding] of Object.entries(
        component.custom_chart.spec.encodings,
      )) {
        const match = /^\{\{ \.params\.(\w+) \}\}$/.exec(encoding.field);
        expect(match, `${entry.chartType}.${channel}`).not.toBeNull();
        expect(declared, `${entry.chartType}.${channel}`).toContain(match![1]);
      }
    }
  });

  it("selects every field the encodings and the ORDER BY reference", () => {
    for (const entry of charts) {
      const component = componentFor(entry);
      const sql = component.custom_chart.metrics_sql;

      const selected = paramRefs(
        /SELECT ([\s\S]*?) FROM /.exec(sql)?.[1] ?? "",
      );
      const referenced = [
        ...paramRefs(JSON.stringify(component.custom_chart.spec)),
        ...paramRefs(sql.slice(sql.indexOf(" FROM "))),
      ].filter((name) => name !== "metrics_view" && name !== "limit");

      for (const name of new Set(referenced)) {
        expect(
          selected,
          `${entry.chartType} does not select ${name}`,
        ).toContain(name);
      }
    }
  });

  it("orders time series chronologically and rankings by the measure", () => {
    // A line chart is bound to a time dimension, so the LIMIT must keep rows in time order,
    // largest series first within each bucket.
    expect(
      componentFor(chartNamed("Line Chart")).custom_chart.metrics_sql,
    ).toMatch(
      /ORDER BY \{\{ \.params\.x_axis \}\}, \{\{ \.params\.y_axis \}\} DESC LIMIT/,
    );

    // A bar chart is a ranking, so the LIMIT must keep the largest rows.
    expect(
      componentFor(chartNamed("Bar Chart")).custom_chart.metrics_sql,
    ).toMatch(/ORDER BY .* DESC LIMIT/);
  });

  it("never emits SQL the metrics dialect rejects", () => {
    for (const entry of charts) {
      const sql = componentFor(entry).custom_chart.metrics_sql;
      expect(sql, entry.chartType).not.toMatch(/\bCASE\b|\{\{-|-\}\}|\bAS\b/i);
    }
  });

  it("selects an unencoded detail dimension for row-level charts", () => {
    for (const entry of charts) {
      const component = componentFor(entry);
      const sql = component.custom_chart.metrics_sql;
      const declaresDetail = component.params.some((p) => p.name === "detail");
      expect(declaresDetail, entry.chartType).toBe(entry.rowLevel);
      expect(sql.includes("{{ .params.detail }}"), entry.chartType).toBe(
        entry.rowLevel,
      );
      expect(
        JSON.stringify(component.custom_chart.spec),
        entry.chartType,
      ).not.toContain(".params.detail");
    }
    expect(chartNamed("Histogram").rowLevel).toBe(true);
    expect(chartNamed("Bar Chart").rowLevel).toBe(false);
  });

  it("only lists examples a metrics view can bind, with one time dimension at most", () => {
    for (const entry of charts) {
      const timeBindings = entry.bindings.filter(
        (binding) => binding.role === "time_dimension",
      );
      expect(timeBindings.length, entry.name).toBeLessThanOrEqual(1);
    }
  });

  it("gives every donut an inner radius, without which Flint draws a pie", () => {
    for (const entry of charts.filter(
      (chart) => chart.chartType === "Donut Chart",
    )) {
      expect(entry.chartProperties?.innerRadius, entry.name).toBeTruthy();
    }
  });

  it("sizes the default row limit to the chart's shape", () => {
    const limitOf = (name: string) =>
      charts.find((chart) => chart.name === name)?.defaultLimit;
    // One mark per category.
    expect(chartNamed("Bar Chart").defaultLimit).toBe(25);
    expect(chartNamed("Pie Chart").defaultLimit).toBe(10);
    // A row per category per series, so a dozen of each fit before the cut.
    expect(
      limitOf("stacked_bar_chart__electricity_generation_mix_by_country_2023"),
    ).toBe(150);
    // Daily buckets over a few months, with and without series.
    expect(limitOf("line_chart__flint_line_over_survey_waves_by_region")).toBe(
      1000,
    );
    expect(limitOf("line_chart__us_unemployment_rate_2000_2023")).toBe(500);
    expect(chartNamed("Candlestick Chart").defaultLimit).toBe(90);
    // One row per observation: fewer where each is drawn as a point.
    expect(chartNamed("Scatter Plot").defaultLimit).toBe(500);
    expect(chartNamed("Histogram").defaultLimit).toBe(1000);
  });

  it("never emits an encoding type, since Rill supplies the semantics", () => {
    for (const entry of charts) {
      const component = componentFor(entry);
      for (const [channel, encoding] of Object.entries(
        component.custom_chart.spec.encodings,
      )) {
        expect(
          encoding,
          `${entry.chartType}.${channel} hardcodes a type`,
        ).not.toHaveProperty("type");
      }
    }
  });

  it("keeps the templated SQL on one line so the YAML stays readable", () => {
    for (const entry of charts) {
      const sqlLine = exampleComponentYAML(entry)
        .split("\n")
        .find((line) => line.includes("metrics_sql"));
      expect(sqlLine, entry.chartType).toContain("SELECT");
      expect(sqlLine, entry.chartType).toContain("LIMIT");
    }
  });
});
