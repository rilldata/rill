import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  computePercentOfTotal,
  updateFilterOnSearch,
} from "./dimension-table-utils";
import { describe, it, expect } from "vitest";

const emptyFilter = {
  include: [],
  exclude: [],
};

const filterWithDimension = {
  include: [{ name: "fruit", in: ["banana", "grapes"] }],
  exclude: [],
};

describe("updateFilterOnSearch", () => {
  it("should return the filter set with search text for an empty filter", () => {
    const updatedfilter = updateFilterOnSearch(emptyFilter, "apple", "fruit");
    expect(updatedfilter).toEqual({
      include: [{ name: "fruit", in: [], like: ["%apple%"] }],
      exclude: [],
    });
  });
  it("should return the filter set with search text for an existing filter", () => {
    const updatedfilter = updateFilterOnSearch(
      filterWithDimension,
      "apple",
      "fruit",
    );
    expect(updatedfilter).toEqual({
      include: [{ name: "fruit", in: ["banana", "grapes"], like: ["%apple%"] }],
      exclude: [],
    });
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
