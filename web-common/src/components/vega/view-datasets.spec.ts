import * as vega from "vega";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import { stubCanvasContext } from "../../features/components/charts/cartesian/test-fixtures";
import { pickViewDatasets } from "./view-datasets";

const rows = [
  { category: "a", value: 1 },
  { category: "b", value: 2 },
];

function headlessView(spec: TopLevelSpec) {
  stubCanvasContext();
  return new vega.View(vega.parse(compile(spec).spec), { renderer: "none" });
}

describe("pickViewDatasets", () => {
  it("keeps the datasets the view defines and drops the rest", () => {
    const view = headlessView({
      data: { name: "metrics-view" },
      mark: "bar",
      encoding: {
        x: { field: "category", type: "nominal" },
        y: { field: "value", type: "quantitative" },
      },
    });

    expect(
      pickViewDatasets(view, { "metrics-view": rows, query1: rows }),
    ).toEqual({ "metrics-view": rows });
  });

  it("drops a named source that the compiled spec no longer has", () => {
    // Vega-Lite drops a named source no mark reads, so declaring it is not
    // enough: this is the shape a donut with an "Other" slice used to compile to.
    const view = headlessView({
      data: { name: "metrics-view" },
      layer: [
        {
          data: { values: rows },
          mark: "arc",
          encoding: {
            theta: { field: "value", type: "quantitative" },
            color: { field: "category", type: "nominal" },
          },
        },
      ],
    });

    expect(pickViewDatasets(view, { "metrics-view": rows })).toEqual({});
    // The failure the filter prevents: svelte-vega would have made this call.
    expect(() =>
      view.change("metrics-view", vega.changeset().insert(rows)),
    ).toThrow("Unrecognized data set: metrics-view");
  });
});
