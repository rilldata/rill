import { describe, expect, it } from "vitest";
import { removeMeasureFromComponentSpec } from "./canvas";

describe("removeMeasureFromComponentSpec", () => {
  it("removes the measure from lists and the measure property only", () => {
    expect(
      removeMeasureFromComponentSpec(
        {
          metrics_view: "profit",
          title: "profit",
          measures: ["revenue", "profit"],
          columns: ["profit"],
          measure: "profit",
          dimensions: ["region"],
        },
        "profit",
      ),
    ).toEqual({ measures: ["revenue"], columns: [], measure: undefined });
  });

  it("leaves non-measure lists alone even when an entry equals the name", () => {
    expect(
      removeMeasureFromComponentSpec(
        {
          measures: ["delta"],
          comparison: ["delta", "percent_change"],
          dimensions: ["delta"],
        },
        "delta",
      ),
    ).toEqual({ measures: [] });
  });

  it("cleans chart field configs and keeps a remaining multi-field measure", () => {
    expect(
      removeMeasureFromComponentSpec(
        {
          x: { field: "region", type: "nominal" },
          y: {
            field: "profit",
            type: "quantitative",
            fields: ["profit", "revenue"],
          },
          color: { field: "profit", type: "quantitative" },
          size: { field: "revenue", type: "quantitative" },
        },
        "profit",
      ),
    ).toEqual({
      y: { field: "revenue", type: "quantitative", fields: ["revenue"] },
      color: undefined,
    });
  });

  it("drops per-measure formatting entries", () => {
    expect(
      removeMeasureFromComponentSpec(
        {
          measures: ["profit"],
          conditional_format: [
            { measure: "profit", mode: "heatmap" },
            { measure: "revenue", mode: "data_bar" },
          ],
        },
        "profit",
      ),
    ).toEqual({
      measures: [],
      conditional_format: [{ measure: "revenue", mode: "data_bar" }],
    });
  });

  it("returns nothing when the measure is unused", () => {
    expect(
      removeMeasureFromComponentSpec(
        { measures: ["revenue"], y: { field: "revenue" } },
        "profit",
      ),
    ).toEqual({});
  });
});
