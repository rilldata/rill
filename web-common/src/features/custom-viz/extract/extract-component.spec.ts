import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";
import { extractedComponentYAML } from "./extract-component";

const spec = {
  metrics_sql: [
    "SELECT country, total_bids FROM auction_metrics WHERE country != 'US' ORDER BY total_bids DESC",
  ],
  vega_spec:
    '{"data": {"name": "query1"}, "mark": "bar", "encoding": {"x": {"field": "country"}, "y": {"field": "total_bids"}}}',
};

describe("extractedComponentYAML", () => {
  it("always lifts the metrics view via the FROM clause", () => {
    const { yaml, bindings } = extractedComponentYAML(
      spec,
      "auction_metrics",
      "My Viz",
      [],
    );

    const parsed = parseDocument(yaml).toJSON();
    expect(parsed.type).toBe("component");
    expect(parsed.display_name).toBe("My Viz");
    expect(parsed.params).toEqual([
      { name: "metrics_view", type: "metrics_view", required: true },
    ]);
    expect(parsed.custom_chart.metrics_sql[0]).toContain(
      "FROM {{ .params.metrics_view }}",
    );
    // Field names stay hardcoded when not lifted.
    expect(parsed.custom_chart.metrics_sql[0]).toContain("total_bids");
    expect(parsed.custom_chart.vega_spec).toContain('"field": "country"');

    expect(bindings).toEqual({ metrics_view: "auction_metrics" });
  });

  it("lifts checked fields in both the metrics SQL and the vega spec", () => {
    const { yaml, bindings } = extractedComponentYAML(
      spec,
      "auction_metrics",
      "My Viz",
      [
        { name: "total_bids", type: "measure" },
        { name: "country", type: "dimension" },
      ],
    );

    const parsed = parseDocument(yaml).toJSON();
    expect(parsed.params).toEqual([
      { name: "metrics_view", type: "metrics_view", required: true },
      { name: "total_bids", type: "measure", required: true },
      { name: "country", type: "dimension", required: true },
    ]);

    const sql = parsed.custom_chart.metrics_sql[0];
    expect(sql).toContain("SELECT {{ .params.country }}");
    expect(sql).toContain("{{ .params.total_bids }} DESC");
    expect(parsed.custom_chart.vega_spec).toContain(
      '"field": "{{ .params.country }}"',
    );
    expect(parsed.custom_chart.vega_spec).toContain(
      '"field": "{{ .params.total_bids }}"',
    );

    expect(bindings).toEqual({
      metrics_view: "auction_metrics",
      total_bids: "total_bids",
      country: "country",
    });
  });

  it("declares time dimensions as time_dimension params", () => {
    const timeSpec = {
      metrics_sql: ["SELECT ts, total_bids FROM auction_metrics"],
      vega_spec: '{"encoding": {"x": {"field": "ts", "type": "temporal"}}}',
    };
    const { yaml } = extractedComponentYAML(timeSpec, "auction_metrics", "V", [
      { name: "ts", type: "dimension", isTimeDimension: true },
    ]);
    const parsed = parseDocument(yaml).toJSON();
    expect(parsed.params).toContainEqual({
      name: "ts",
      type: "time_dimension",
      required: true,
    });
  });

  it("does not replace substrings of other tokens", () => {
    const trickySpec = {
      metrics_sql: ["SELECT bid, total_bid FROM mv"],
      vega_spec: '{"field": "total_bid"}',
    };
    const { yaml } = extractedComponentYAML(trickySpec, "mv", "V", [
      { name: "bid", type: "measure" },
    ]);
    const parsed = parseDocument(yaml).toJSON();
    // "bid" was lifted, but "total_bid" (which contains "bid" as a substring
    // separated by an underscore, i.e. no word boundary) stays untouched.
    expect(parsed.custom_chart.metrics_sql[0]).toBe(
      "SELECT {{ .params.bid }}, total_bid FROM {{ .params.metrics_view }}",
    );
    expect(parsed.custom_chart.vega_spec).toBe('{"field": "total_bid"}');
  });

  it("sanitizes and dedupes generated param names", () => {
    const trickySpec = {
      metrics_sql: ["SELECT `weird field`, metrics_view FROM mv"],
      vega_spec: "{}",
    };
    const { yaml } = extractedComponentYAML(trickySpec, "mv", "V", [
      { name: "weird field", type: "dimension" },
      { name: "metrics_view", type: "measure" },
    ]);
    const parsed = parseDocument(yaml).toJSON();
    const names = parsed.params.map((p: { name: string }) => p.name);
    expect(names).toContain("weird_field");
    // Collides with the always-present metrics_view param, so it gets a suffix.
    expect(names).toContain("metrics_view_2");
  });

  it("produces YAML that round-trips through a parser", () => {
    const { yaml } = extractedComponentYAML(spec, "auction_metrics", "V", [
      { name: "country", type: "dimension" },
    ]);
    const doc = parseDocument(yaml);
    expect(doc.errors).toEqual([]);
  });
});
