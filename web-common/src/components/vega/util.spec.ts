import { describe, expect, it } from "vitest";
import { parseExpression } from "vega-expression";
import { sanitizeFieldName } from "./util";

describe("sanitizeFieldName", () => {
  it("keeps simple field names readable", () => {
    expect(sanitizeFieldName("total_sales")).toBe("rill_total_sales");
  });

  it("returns a valid Vega expression function name for measure names with operators", () => {
    const measureNames = [
      "Total Sample Revenue",
      "Sample Rate* Lift",
      "Share(%) | Variant A",
      "Share(%) | Baseline",
      "Avg Sample Value | Variant A",
      "Avg Sample Value | Baseline",
      "Success Rate | Variant A",
      "Success Rate | Baseline",
      "Success Rate | Delta",
    ];

    for (const measureName of measureNames) {
      const formatType = sanitizeFieldName(measureName);

      expect(formatType).toMatch(/^[A-Za-z_$][A-Za-z0-9_$]*$/);
      expect(() => parseExpression(`${formatType}(datum.value)`)).not.toThrow();
    }
  });
});
