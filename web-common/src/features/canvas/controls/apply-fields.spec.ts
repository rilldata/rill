import { describe, expect, it } from "vitest";
import { applyFieldControls } from "./apply-fields";
import type { AppliedControl } from "./controls";

function applied(
  control: AppliedControl["control"],
  values: string[] | null,
): AppliedControl {
  return {
    control,
    values,
    level: "item",
    scope: ["chart"],
    urlKey: "c.chart.x",
    targetKey: `${control.kind}:${control.target}`,
  };
}

const spec = {
  metrics_view: "spend",
  title: "spend",
  x: { field: "device" },
  y: { field: "spend", fields: ["spend", "clicks"] },
  color: { field: "device" },
  columns: ["device", { name: "spend", width: 10 }],
  sort_by: "spend",
};

describe("applyFieldControls", () => {
  it("returns the same spec when nothing is selected", () => {
    const swap = applied({ kind: "measure", target: "spend" }, ["spend"]);
    expect(applyFieldControls(spec, [])).toBe(spec);
    expect(applyFieldControls(spec, [swap])).toBe(spec);
  });

  it("swaps a field wherever it is used as a field, and nowhere else", () => {
    const swap = applied({ kind: "measure", target: "spend" }, ["impressions"]);
    expect(applyFieldControls(spec, [swap])).toEqual({
      metrics_view: "spend",
      title: "spend",
      x: { field: "device" },
      y: { field: "impressions", fields: ["impressions", "clicks"] },
      color: { field: "device" },
      columns: ["device", { name: "impressions", width: 10 }],
      sort_by: "impressions",
    });
    expect(spec.sort_by).toBe("spend");
  });

  it("applies several swaps in one pass", () => {
    const a = applied({ kind: "measure", target: "spend" }, ["clicks"]);
    const b = applied({ kind: "measure", target: "clicks" }, ["spend"]);
    const result = applyFieldControls(spec, [a, b]);
    expect(result.y.fields).toEqual(["clicks", "spend"]);
  });

  it("sets a slot to one field or a list", () => {
    const one = applied({ kind: "slot", target: "color.field" }, ["country"]);
    expect(applyFieldControls(spec, [one]).color).toEqual({ field: "country" });

    const many = applied({ kind: "slot", target: "columns", multiple: true }, [
      "country",
      "clicks",
    ]);
    expect(applyFieldControls(spec, [many]).columns).toEqual([
      "country",
      "clicks",
    ]);

    const unset = applied({ kind: "slot", target: "color.field" }, null);
    expect(applyFieldControls(spec, [unset])).toBe(spec);
  });
});
