import { describe, expect, it } from "vitest";
import { metricsSQLUsesTimeRange } from "./expression-utils";

describe("metricsSQLUsesTimeRange", () => {
  const timeless = new Map([["products", false]]);
  const mixed = new Map([
    ["products", false],
    ["orders", true],
  ]);

  it.each([
    {
      name: "a dashboard without a time dimension never resolves a time range",
      views: ["products"],
      dashboard: timeless,
      expected: false,
    },
    {
      name: "a metrics view without a time dimension is not filtered by another one's",
      views: ["products"],
      dashboard: mixed,
      expected: false,
    },
    {
      name: "one query on a metrics view with a time dimension is enough",
      views: ["products", "orders"],
      dashboard: mixed,
      expected: true,
    },
    {
      name: "an unknown metrics view follows a dashboard with a time dimension",
      views: [undefined],
      dashboard: mixed,
      expected: true,
    },
    {
      name: "an unknown metrics view does not wait on a dashboard without a time dimension",
      views: [undefined, "missing"],
      dashboard: timeless,
      expected: false,
    },
  ])("$name", ({ views, dashboard, expected }) => {
    expect(metricsSQLUsesTimeRange(views, dashboard)).toBe(expected);
  });
});
