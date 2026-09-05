import { describe, it, expect } from "vitest";
import type { V1MetricsViewAggregationMeasure } from "@rilldata/web-common/runtime-client";
import { prepareMeasuresForRequest } from "../pivot/pivot-utils";
import type { EphemeralMeasureDef } from "./types";
import {
  slugifyEphemeralMeasureName,
  validateEphemeralMeasureDef,
  validateEphemeralMeasureName,
} from "./validation";

const PROFIT: EphemeralMeasureDef = {
  name: "profit",
  displayName: "Profit",
  expression: "revenue - cost",
};

describe("prepareMeasuresForRequest", () => {
  it("attaches the expression compute to ephemeral measures", () => {
    const measures: V1MetricsViewAggregationMeasure[] = [
      { name: "revenue" },
      { name: "profit" },
    ];
    expect(prepareMeasuresForRequest(measures, [PROFIT])).toEqual([
      { name: "revenue" },
      {
        name: "profit",
        expression: { expression: "revenue - cost", displayName: "Profit" },
      },
    ]);
  });

  it("still maps comparison suffixes for spec measures", () => {
    const measures: V1MetricsViewAggregationMeasure[] = [
      { name: "revenue" },
      { name: "revenue__delta_abs" },
      { name: "revenue__delta_rel" },
      { name: "profit" },
    ];
    expect(prepareMeasuresForRequest(measures, [PROFIT])).toEqual([
      { name: "revenue" },
      { name: "revenue__delta_abs", comparisonDelta: { measure: "revenue" } },
      { name: "revenue__delta_rel", comparisonRatio: { measure: "revenue" } },
      {
        name: "profit",
        expression: { expression: "revenue - cost", displayName: "Profit" },
      },
    ]);
  });

  it("is a no-op without ephemeral measures", () => {
    const measures: V1MetricsViewAggregationMeasure[] = [{ name: "revenue" }];
    expect(prepareMeasuresForRequest(measures, undefined)).toEqual(measures);
  });
});

describe("validateEphemeralMeasureName", () => {
  const reserved = new Set(["revenue", "domain"]);

  it.each([
    ["profit", undefined],
    ["Profit_2", undefined],
    ["2profit", "must start with a letter"],
    ["pro fit", "must start with a letter"],
    ["profit__delta_abs", "comparison suffix"],
    ["profit__delta_rel", "comparison suffix"],
    ["profit_prev", "comparison suffix"],
    ["profit_delta", "comparison suffix"],
    ["profit_delta_perc", "comparison suffix"],
    ["profit_percent_of_total", "comparison suffix"],
    ["profit_rill_day", '"_rill_"'],
    ["revenue", "already used"],
  ])("%s", (name, expected) => {
    const error = validateEphemeralMeasureName(name, reserved);
    if (expected === undefined) {
      expect(error).toBeUndefined();
    } else {
      expect(error).toContain(expected);
    }
  });
});

describe("validateEphemeralMeasureDef", () => {
  const known = new Set(["revenue", "cost"]);
  const reserved = new Set(["revenue", "cost", "domain", "timestamp"]);

  it("accepts a valid definition", () => {
    expect(
      validateEphemeralMeasureDef(PROFIT, known, reserved),
    ).toBeUndefined();
  });

  it("rejects unknown measure references", () => {
    const def = { ...PROFIT, expression: "revenue - expenses" };
    expect(validateEphemeralMeasureDef(def, known, reserved)).toContain(
      '"expenses" is not a measure',
    );
  });

  it("rejects references to other ephemeral measures", () => {
    // "profit2" is not in the known measure set, so a second ephemeral
    // measure cannot reference the first.
    const def = {
      name: "margin",
      displayName: "Margin",
      expression: "profit2 / revenue",
    };
    expect(validateEphemeralMeasureDef(def, known, reserved)).toContain(
      '"profit2" is not a measure',
    );
  });

  it("rejects invalid expressions", () => {
    const def = { ...PROFIT, expression: "sum(revenue)" };
    expect(validateEphemeralMeasureDef(def, known, reserved)).toContain(
      "unsupported function",
    );
  });

  it("rejects an empty display name", () => {
    const def = { ...PROFIT, displayName: " " };
    expect(validateEphemeralMeasureDef(def, known, reserved)).toContain(
      "display name is required",
    );
  });
});

describe("slugifyEphemeralMeasureName", () => {
  const reserved = new Set(["profit", "revenue", "cost"]);

  it("derives a plain alias from the display name", () => {
    expect(slugifyEphemeralMeasureName("Gross Margin", reserved)).toBe(
      "gross_margin",
    );
  });

  it("disambiguates collisions with existing fields", () => {
    expect(slugifyEphemeralMeasureName("Profit", reserved)).toBe("profit_2");
  });

  it("always yields an alias that passes name validation", () => {
    for (const label of [
      "Revenue Delta",
      "Cost Prev",
      "Revenue Percent Of Total",
      "Total Rill Value",
      "a__previous",
      "123 abc",
      "Profit",
    ]) {
      const slug = slugifyEphemeralMeasureName(label, reserved);
      expect(validateEphemeralMeasureName(slug, reserved)).toBeUndefined();
    }
    expect(slugifyEphemeralMeasureName("Revenue Delta", reserved)).toBe(
      "revenue_delta_2",
    );
    expect(slugifyEphemeralMeasureName("123 abc", reserved)).toBe("m_123_abc");
  });
});
