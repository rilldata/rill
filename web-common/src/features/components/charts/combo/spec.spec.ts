import { describe, expect, it } from "vitest";
import { at, chartData } from "../cartesian/test-fixtures";
import type { ComboChartSpec } from "./ComboChartProvider";
import { generateVLComboChartSpec } from "./spec";

describe("generateVLComboChartSpec with dotted field names", () => {
  const dimension = "event.time";
  const measure = "stats.count";

  function data() {
    const result = chartData();
    result.fields[dimension] = { name: dimension, displayName: "Event Time" };
    result.fields[measure] = { name: measure, displayName: "Count" };
    result.domainValues = { [dimension]: ["Other", "Scheduled Report"] };
    return result;
  }

  const config: ComboChartSpec = {
    metrics_view: "druid_requests",
    x: { field: dimension, type: "nominal", sort: "-y" },
    y1: { field: measure, type: "quantitative", mark: "bar" },
    y2: { field: "comment_count", type: "quantitative", mark: "line" },
  };

  it("escapes the x field once in the hover layer and keeps the custom sort", () => {
    // The hover helper escapes the x field itself; a pre-escaped name would
    // be escaped twice and the hover rule would reference a missing field.
    const spec = generateVLComboChartSpec(config, data());
    expect(at(spec, "layer.0.encoding.x.field")).toBe("event\\.time");
    expect(at(spec, "layer.0.encoding.x.sort")).toEqual([
      "Other",
      "Scheduled Report",
    ]);
  });

  it("escapes the measures handed to fold and filters on the raw names", () => {
    // Vega emits the raw measure name as the fold key, so the per-measure
    // filters and the color domain keep the raw names.
    const spec = generateVLComboChartSpec(config, data());
    expect(at(spec, "transform.0.fold")).toEqual([
      "stats\\.count",
      "comment_count",
    ]);
    expect(at(spec, "layer.1.transform.0.filter")).toBe(
      "datum['Measure'] === 'stats.count'",
    );
    expect(at(spec, "layer.1.encoding.color.scale.domain")).toEqual([
      "stats.count",
      "comment_count",
    ]);
  });
});
