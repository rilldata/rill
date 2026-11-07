import { describe, expect, it } from "vitest";
import { d3ToExcelFormat, excelNumberFormat } from "./number-format";

describe("excelNumberFormat", () => {
  it("maps format presets", () => {
    expect(excelNumberFormat({ formatPreset: "currency_usd" })).toBe(
      `"$"#,##0.00`,
    );
    expect(excelNumberFormat({ formatPreset: "currency_eur" })).toBe(
      `"€"#,##0.00`,
    );
    expect(excelNumberFormat({ formatPreset: "percentage" })).toBe("0.0%");
    expect(excelNumberFormat({ formatPreset: "none" })).toBe("General");
    expect(excelNumberFormat({ formatPreset: "interval_ms" })).toBe("General");
  });

  it("picks a grouped integer or decimal format from the values", () => {
    expect(
      excelNumberFormat({ formatPreset: "humanize" }, [1, 2000, null]),
    ).toBe("#,##0");
    expect(excelNumberFormat({ formatPreset: "humanize" }, [1.5, 2])).toBe(
      "#,##0.00",
    );
    expect(excelNumberFormat(undefined)).toBe("#,##0");
  });

  it("prefers a d3 format when the measure has one", () => {
    expect(excelNumberFormat({ formatD3: "$,.2f" })).toBe(`"$"#,##0.00`);
  });

  it("takes the currency symbols from the measure's d3 locale", () => {
    expect(
      excelNumberFormat({
        formatD3: "$,.2f",
        formatD3Locale: { currency: ["£", ""] },
      }),
    ).toBe(`"£"#,##0.00`);
    expect(
      excelNumberFormat({
        formatD3: "$,.0f",
        formatD3Locale: { currency: ["", " €"] },
      }),
    ).toBe(`#,##0" €"`);
    // A locale without a currency keeps d3's dollar.
    expect(
      excelNumberFormat({
        formatD3: "$,.2f",
        formatD3Locale: { decimal: "," },
      }),
    ).toBe(`"$"#,##0.00`);
  });
});

describe("d3ToExcelFormat", () => {
  it("translates the common specifiers", () => {
    expect(d3ToExcelFormat(",")).toBe("#,##0");
    expect(d3ToExcelFormat(",", [1.25])).toBe("#,##0.00");
    expect(d3ToExcelFormat(".0f")).toBe("0");
    expect(d3ToExcelFormat(",.1f")).toBe("#,##0.0");
    expect(d3ToExcelFormat(".2f")).toBe("0.00");
    expect(d3ToExcelFormat("$,.2f")).toBe(`"$"#,##0.00`);
    expect(d3ToExcelFormat(".1%")).toBe("0.0%");
    expect(d3ToExcelFormat("%")).toBe("0%");
    expect(d3ToExcelFormat(",d")).toBe("#,##0");
    expect(d3ToExcelFormat(".3s", [12345])).toBe("0.000");
  });

  it("rejects specifiers it can't express", () => {
    expect(d3ToExcelFormat("not a format")).toBeUndefined();
    expect(d3ToExcelFormat(".2b")).toBeUndefined();
  });
});
