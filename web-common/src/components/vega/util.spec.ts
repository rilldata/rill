import { describe, expect, it } from "vitest";
import { parseExpression } from "vega-expression";
import { sanitizeFieldName } from "./util";

describe("sanitizeFieldName", () => {
  it("keeps simple field names readable", () => {
    expect(sanitizeFieldName("total_sales")).toBe("rill_total_sales");
  });

  it("returns a valid Vega expression function name for measure names with operators", () => {
    const measureNames: [string, string][] = [
      ["Total Sample Revenue", "rill_Total_u20_Sample_u20_Revenue"],
      ["Sample Rate* Lift", "rill_Sample_u20_Rate_u2a__u20_Lift"],
      [
        "Share(%) | Variant A",
        "rill_Share_u28__u25__u29__u20__u7c__u20_Variant_u20_A",
      ],
      [
        "Share(%) | Baseline",
        "rill_Share_u28__u25__u29__u20__u7c__u20_Baseline",
      ],
      [
        "Avg Sample Value | Variant A",
        "rill_Avg_u20_Sample_u20_Value_u20__u7c__u20_Variant_u20_A",
      ],
      [
        "Avg Sample Value | Baseline",
        "rill_Avg_u20_Sample_u20_Value_u20__u7c__u20_Baseline",
      ],
      [
        "Success Rate | Variant A",
        "rill_Success_u20_Rate_u20__u7c__u20_Variant_u20_A",
      ],
      [
        "Success Rate | Baseline",
        "rill_Success_u20_Rate_u20__u7c__u20_Baseline",
      ],
      ["Success Rate | Delta", "rill_Success_u20_Rate_u20__u7c__u20_Delta"],
    ];

    for (const [measureName, expectedFormatType] of measureNames) {
      const formatType = sanitizeFieldName(measureName);

      expect(formatType).toBe(expectedFormatType);
      expect(() => parseExpression(`${formatType}(datum.value)`)).not.toThrow();
    }
  });
});
