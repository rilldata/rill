import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  computePercentOfTotal,
  estimateColumnSizes,
  updateFilterOnSearch,
} from "./dimension-table-utils";
import { describe, it, expect } from "vitest";
import { DIMENSION_TABLE_CONFIG } from "./DimensionTableConfig";

describe("updateFilterOnSearch", () => {
  it("should return the filter set with search text for an empty filter", () => {
    const updatedfilter = updateFilterOnSearch(
      createAndExpression([]),
      "apple",
      "fruit",
    );
    expect(updatedfilter).toEqual(
      createAndExpression([createLikeExpression("fruit", "%apple%")]),
    );
  });
  it("should return the filter set with search text for an existing filter", () => {
    const updatedfilter = updateFilterOnSearch(
      createAndExpression([createInExpression("fruit", ["banana", "grapes"])]),
      "apple",
      "fruit",
    );
    expect(updatedfilter).toEqual(
      createAndExpression([
        createLikeExpression("fruit", "%apple%"),
        createInExpression("fruit", ["banana", "grapes"]),
      ]),
    );
  });
});

const expectedPOTData = [
  {
    fruit: "banana",
    measure_0: 20,
    measure_0_percent_of_total: {
      int: "20",
      percent: "%",
      dot: "",
      frac: "",
      suffix: "",
    },
  },
  {
    fruit: "grapes",
    measure_0: 15,
    measure_0_percent_of_total: {
      int: "15",
      percent: "%",
      dot: "",
      frac: "",
      suffix: "",
    },
  },
  {
    fruit: "oranges",
    measure_0: 25,
    measure_0_percent_of_total: {
      int: "25",
      percent: "%",
      dot: "",
      frac: "",
      suffix: "",
    },
  },
  {
    fruit: "apple",
    measure_0: 30,
    measure_0_percent_of_total: {
      int: "30",
      percent: "%",
      dot: "",
      frac: "",
      suffix: "",
    },
  },
  {
    fruit: "guvava",
    measure_0: 35,
    measure_0_percent_of_total: {
      int: "35",
      percent: "%",
      dot: "",
      frac: "",
      suffix: "",
    },
  },
];

describe("computePercentOfTotal", () => {
  const values = [
    { fruit: "banana", measure_0: 20 },
    { fruit: "grapes", measure_0: 15 },
    { fruit: "oranges", measure_0: 25 },
    { fruit: "apple", measure_0: 30 },
    { fruit: "guvava", measure_0: 35 },
  ];

  it("should compute % of total correctly with non-zero total", () => {
    const total = 100;
    const computedValues = computePercentOfTotal(values, total, "measure_0");
    expect(computedValues).toEqual(expectedPOTData);
  });
});

describe("computePercentOfTotal", () => {
  const values = [
    { fruit: "banana", measure_0: 20 },
    { fruit: "grapes", measure_0: 15 },
    { fruit: "oranges", measure_0: 25 },
    { fruit: "apple", measure_0: 30 },
    { fruit: "guvava", measure_0: 35 },
  ];

  it("should compute % of total correctly with zero total", () => {
    const total = 0;
    const computedValues = computePercentOfTotal(values, total, "measure_0");

    const expected = values.map((value) => ({
      ...value,
      measure_0_percent_of_total: PERC_DIFF.CURRENT_VALUE_NO_DATA,
    }));

    expect(computedValues).toEqual(expected);
  });
});

describe("estimateColumnSizes", () => {
  it("expands comparison columns to fit longer formatted values", () => {
    const columns = [
      { name: "org", type: "VARCHAR" },
      { name: "cost_delta", type: "RILL_CHANGE" },
      { name: "cost", type: "INT" },
    ];

    const sizes = estimateColumnSizes(
      columns,
      { org: 8, cost_delta: 12, cost: 5 },
      1200,
      { ...DIMENSION_TABLE_CONFIG },
    );

    expect(sizes[1]).toBeGreaterThan(DIMENSION_TABLE_CONFIG.comparisonColumnWidth);
  });

  it("keeps comparison columns at the default minimum when values are short", () => {
    const columns = [
      { name: "org", type: "VARCHAR" },
      { name: "cost_delta", type: "RILL_CHANGE" },
      { name: "cost", type: "INT" },
    ];

    const sizes = estimateColumnSizes(
      columns,
      { org: 8, cost_delta: 2, cost: 5 },
      1200,
      { ...DIMENSION_TABLE_CONFIG },
    );

    expect(sizes[1]).toBe(DIMENSION_TABLE_CONFIG.comparisonColumnWidth);
  });
});
