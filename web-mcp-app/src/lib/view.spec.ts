import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/bar_chart.json";
import empty from "../../harness/fixtures/empty.json";
import noData from "../../harness/fixtures/no_data.json";
import toolError from "../../harness/fixtures/tool_error.json";
import unsupported from "../../harness/fixtures/unsupported_widget.json";
import { interpretToolResult } from "./view";

describe("interpretToolResult", () => {
  it("returns the chart for a complete result", () => {
    const interpreted = interpretToolResult(bar.result as never);
    expect(interpreted.kind).toBe("chart");
  });

  it("returns the chart for a zero-row result, whose data is an empty array", () => {
    expect(empty.result.structuredContent.data).toEqual([]);
    const interpreted = interpretToolResult(empty.result as never);
    expect(interpreted.kind).toBe("chart");
  });

  it("explains a result without data and keeps the deep link", () => {
    const interpreted = interpretToolResult(noData.result as never);
    expect(interpreted).toEqual({
      kind: "error",
      message: expect.stringContaining("did not include chart data"),
      openUrl: noData.result.structuredContent.open_url,
    });
  });

  it("explains unsupported widget modes", () => {
    const interpreted = interpretToolResult(unsupported.result as never);
    expect(interpreted.kind).toBe("error");
    if (interpreted.kind === "error") {
      expect(interpreted.message).toContain("dashboard");
      expect(interpreted.openUrl).toContain("canvas/overview");
    }
  });

  it("surfaces tool errors with their text", () => {
    const interpreted = interpretToolResult(toolError.result as never);
    expect(interpreted).toEqual({
      kind: "error",
      message: 'metrics view "missing_metrics" not found',
    });
  });
});
