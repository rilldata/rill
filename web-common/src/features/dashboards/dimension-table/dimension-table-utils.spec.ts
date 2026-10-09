import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  computePercentOfTotal,
  getDimensionFilterWithSearch,
  updateFilterOnSearch,
} from "./dimension-table-utils";
import { describe, it, expect } from "vitest";

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

describe("getDimensionFilterWithSearch", () => {
  it("should apply the search text when there are no active filters", () => {
    const updatedFilter = getDimensionFilterWithSearch(
      undefined,
      "apple",
      "fruit",
    );
    expect(updatedFilter).toEqual(
      createAndExpression([createLikeExpression("fruit", "%apple%")]),
    );
  });
  it("should return an empty filter when there are no active filters and no search text", () => {
    const updatedFilter = getDimensionFilterWithSearch(undefined, "", "fruit");
    expect(updatedFilter).toEqual(createAndExpression([]));
  });
  it("should replace the filter on the searched dimension and keep the others", () => {
    const updatedFilter = getDimensionFilterWithSearch(
      createAndExpression([
        createInExpression("fruit", ["banana"]),
        createInExpression("color", ["red"]),
      ]),
      "apple",
      "fruit",
    );
    expect(updatedFilter).toEqual(
      createAndExpression([
        createInExpression("color", ["red"]),
        createLikeExpression("fruit", "%apple%"),
      ]),
    );
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
