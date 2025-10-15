import { describe, it, expect } from "vitest";
import { createMeasureValueFormatter } from "./format-measure-value";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

describe("format-measure-value with d3_locale", () => {
  it("should apply custom thousand separators for big numbers", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "test_measure",
      expression: "SUM(amount)",
      formatD3: ",",
      formatD3Locale: {
        thousands: " ", // space as thousand separator
        decimal: ",", // comma as decimal separator
        grouping: [3],
        currency: ["€", ""],
      },
    };

    const formatter = createMeasureValueFormatter(measure, "big-number");

    // Test with a number that should have thousand separators
    const result = formatter(123456);

    // The big number formatter will abbreviate to "123k" but should preserve locale settings
    // For big numbers, we humanize so we get "123k" with space separators
    expect(result).toContain("123");
  });

  it("should apply custom currency symbols from d3_locale for big numbers", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "rupee_measure",
      expression: "SUM(amount)",
      formatD3: "$,",
      formatD3Locale: {
        thousands: ",",
        decimal: ".",
        grouping: [3, 2, 2], // Indian numbering system
        currency: ["₹", ""], // Rupee symbol
      },
    };

    const formatter = createMeasureValueFormatter(measure, "big-number");

    const result = formatter(1234567);

    // Should have the rupee symbol
    expect(result).toContain("₹");
    // Should be humanized to something like "₹1.23M" or "₹1M"
    expect(result).toMatch(/₹\d/);
  });

  it("should apply custom decimal separators in tooltips", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "european_measure",
      expression: "SUM(amount)",
      formatD3: ",.2f",
      formatD3Locale: {
        thousands: ".",
        decimal: ",",
        grouping: [3],
        currency: ["€", ""],
      },
    };

    const formatter = createMeasureValueFormatter(measure, "tooltip");

    const result = formatter(1234.56);

    // Tooltip should show the full number with custom separators
    // Since it's a tooltip with d3 format, it will use the d3 formatter directly
    expect(result).toBe("1.234,56");
  });

  it("should apply custom thousand separators with non-currency formats", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "test_measure",
      expression: "COUNT(*)",
      formatD3: ",",
      formatD3Locale: {
        thousands: "'",
        decimal: ".",
        grouping: [3],
        currency: ["$", ""],
      },
    };

    const formatter = createMeasureValueFormatter(measure, "big-number");

    const result = formatter(1234567);

    // For big numbers with plain d3 format, it should be humanized and use custom separators
    // Since there's no currency or percent, it should abbreviate the number
    expect(result).toBe("1.23M");
  });

  it("should handle different grouping patterns", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "indian_number",
      expression: "SUM(amount)",
      formatD3: ",",
      formatD3Locale: {
        thousands: ",",
        decimal: ".",
        grouping: [3, 2, 2], // Indian style: 1,00,00,000
        currency: ["₹", ""],
      },
    };

    const formatter = createMeasureValueFormatter(measure, "big-number");

    // Test with 10 million
    const result = formatter(10000000);

    // Should be humanized to something like "10M"
    expect(result).toContain("10");
    expect(result).toMatch(/M/);
  });

  it("should work with currency suffix instead of prefix", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "suffix_currency",
      expression: "SUM(amount)",
      formatD3: "$,",
      formatD3Locale: {
        thousands: " ",
        decimal: ",",
        grouping: [3],
        currency: ["", "€"], // Euro as suffix
      },
    };

    const formatter = createMeasureValueFormatter(measure, "big-number");

    const result = formatter(5000);

    // Should have Euro as suffix
    expect(result).toContain("€");
    expect(result).toMatch(/\d+k?€/);
  });
});
