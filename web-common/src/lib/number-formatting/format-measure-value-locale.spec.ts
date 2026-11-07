import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { createMeasureValueFormatter } from "./format-measure-value";

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

    // Big numbers honor the explicit d3 format, including its locale settings
    expect(result).toBe("123 456");
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

    // Should use the rupee symbol and the Indian grouping from the locale
    expect(result).toBe("₹12,34,567");
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

    // Big numbers honor the plain d3 format with the custom separators
    expect(result).toBe("1'234'567");
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

    // Should use the Indian grouping pattern from the locale
    expect(result).toBe("1,00,00,000");
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

    // Should have Euro as suffix and the space thousand separator
    expect(result).toBe("5 000€");
  });
});

describe("format-measure-value with an explicit formatD3", () => {
  // Sub-cent measure, e.g. cost per click: an explicit d3 format must be
  // honored in tooltips and big numbers, otherwise values like $0.0002 get
  // humanized/rounded into invisibility (see the 2-decimal currency default).
  const subCentMeasure: MetricsViewSpecMeasure = {
    name: "cost_per_click",
    expression: "SUM(cost) / SUM(clicks)",
    formatD3: "$,.4f",
  };

  it("honors formatD3 precision in the tooltip context", () => {
    const formatter = createMeasureValueFormatter(subCentMeasure, "tooltip");
    expect(formatter(0.0002)).toBe("$0.0002");
    expect(formatter(0.00023456)).toBe("$0.0002");
    expect(formatter(1234.5678)).toBe("$1,234.5678");
  });

  it("honors formatD3 precision in the big-number context", () => {
    const formatter = createMeasureValueFormatter(subCentMeasure, "big-number");
    expect(formatter(0.0002)).toBe("$0.0002");
    expect(formatter(1234567.8901)).toBe("$1,234,567.8901");
  });

  it("keeps humanized values in the axis context so tick labels stay compact", () => {
    const formatter = createMeasureValueFormatter(subCentMeasure, "axis");
    expect(formatter(0.0002)).toBe("$2e-4");
    expect(formatter(1234567.8901)).toBe("$1.2M");
  });

  it("keeps one decimal on axis ticks so half-unit values are not rounded away", () => {
    const formatter = createMeasureValueFormatter(subCentMeasure, "axis");
    // 1,500,000 previously rounded up to "$2M", making a tick sitting on 1.5M
    // read as 2M on the axis.
    expect(formatter(1_500_000)).toBe("$1.5M");
    expect(formatter(1_000_000)).toBe("$1M");
    expect(formatter(500_000)).toBe("$500k");
    expect(formatter(1_500_000_000)).toBe("$1.5B");
  });

  it("still humanizes preset-formatted measures in the big-number context", () => {
    const presetMeasure: MetricsViewSpecMeasure = {
      name: "total_cost",
      expression: "SUM(cost)",
      formatPreset: "currency_usd",
    };
    const formatter = createMeasureValueFormatter(presetMeasure, "big-number");
    expect(formatter(1234567.8901)).toBe("$1.23M");
  });
});

describe("format-measure-value SI prefix remap (G -> B)", () => {
  const measureWithSI: MetricsViewSpecMeasure = {
    name: "si_measure",
    expression: "SUM(amount)",
    formatD3: ",.3s",
  };

  it("should remap G to B for billions in table context", () => {
    const formatter = createMeasureValueFormatter(measureWithSI, "table");
    expect(formatter(1.5e9)).toBe("1.50B");
  });

  it("should remap G to B for negative billions", () => {
    const formatter = createMeasureValueFormatter(measureWithSI, "table");
    // D3 uses U+2212 (minus sign), not ASCII hyphen, for negatives.
    expect(formatter(-2.5e9)).toBe("−2.50B");
  });

  it("should leave other SI prefixes untouched (M, k, T)", () => {
    const formatter = createMeasureValueFormatter(measureWithSI, "table");
    expect(formatter(1.5e6)).toBe("1.50M");
    expect(formatter(1.5e3)).toBe("1.50k");
    expect(formatter(1.5e12)).toBe("1.50T");
  });

  it("should remap G to B with a currency prefix", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "currency_si",
      expression: "SUM(amount)",
      formatD3: "$,.3s",
    };
    const formatter = createMeasureValueFormatter(measure, "table");
    expect(formatter(1.5e9)).toBe("$1.50B");
  });

  it("should remap G to B with a custom currency from locale", () => {
    const measure: MetricsViewSpecMeasure = {
      name: "rupee_si",
      expression: "SUM(amount)",
      formatD3: "$,.3s",
      formatD3Locale: {
        thousands: ",",
        decimal: ".",
        grouping: [3, 2, 2],
        currency: ["₹", ""],
      },
    };
    const formatter = createMeasureValueFormatter(measure, "table");
    expect(formatter(1.5e9)).toBe("₹1.50B");
  });

  it("should not affect non-SI formats", () => {
    // ",.3f" produces fixed-point output with no SI suffix.
    const measure: MetricsViewSpecMeasure = {
      name: "fixed_measure",
      expression: "SUM(amount)",
      formatD3: ",.3f",
    };
    const formatter = createMeasureValueFormatter(measure, "table");
    expect(formatter(1500000000)).toBe("1,500,000,000.000");
  });
});
