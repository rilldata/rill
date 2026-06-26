import { describe, expect, it } from "vitest";
import { isCanvasExportQuery } from "./settle";

describe("isCanvasExportQuery", () => {
  it("matches query service queries for the active runtime instance", () => {
    expect(
      isCanvasExportQuery(
        ["QueryService", "metricsViewAggregation", "instance-1", {}],
        "instance-1",
      ),
    ).toBe(true);
  });

  it("matches custom chart metrics SQL queries", () => {
    expect(
      isCanvasExportQuery(["metrics_sql", "chart", 0, "select 1", "{}"], "i"),
    ).toBe(true);
  });

  it("ignores unrelated admin and other runtime queries", () => {
    expect(
      isCanvasExportQuery(
        ["AdminService", "getProject", { project: "p" }],
        "instance-1",
      ),
    ).toBe(false);
    expect(
      isCanvasExportQuery(
        ["RuntimeService", "gitStatus", "instance-1", {}],
        "instance-1",
      ),
    ).toBe(false);
    expect(
      isCanvasExportQuery(
        ["QueryService", "metricsViewAggregation", "instance-2", {}],
        "instance-1",
      ),
    ).toBe(false);
  });
});
