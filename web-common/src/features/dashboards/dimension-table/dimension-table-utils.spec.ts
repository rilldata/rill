import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  computeComparisonValues,
  computePercentOfTotal,
  getFilterForComparsion,
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
      "fruit"
    );
    expect(updatedfilter).toEqual({
      include: [{ name: "fruit", in: ["banana", "grapes"], like: ["%apple%"] }],
      exclude: [],
    });
  });
});

const filterValues = ["banana", "grapes", "oranges"];

describe("getFilterForComparsion", () => {
  it("should return the filter set for an empty filter", () => {
    const updatedfilter = getFilterForComparsion(
      emptyFilter,
      "fruit",
      filterValues
    );
    expect(updatedfilter).toEqual({
      include: [{ name: "fruit", in: ["banana", "grapes", "oranges"] }],
      exclude: [],
    });
  });
  it("should return the filter set for an existing filter", () => {
    const updatedfilter = getFilterForComparsion(
      filterWithDimension,
      "fruit",
      filterValues
    );
    expect(updatedfilter).toEqual({
      include: [{ name: "fruit", in: ["banana", "grapes", "oranges"] }],
      exclude: [],
    });
  });
});

const comparisonResponse = {
  meta: [
    {
      name: "fruit",
      type: "CODE_STRING",
      nullable: true,
    },
    {
      name: "measure_0",
      type: "CODE_INT128",
      nullable: true,
    },
  ],

  data: [
    { fruit: "banana", measure_0: 15 },
    { fruit: "grapes", measure_0: 20 },
    { fruit: "oranges", measure_0: 25 },
    { fruit: "guvava", measure_0: 0 },
  ],
};

const expectedData = [
  {
    fruit: "banana",
    measure_0: 20,
    measure_0_delta: 5,
    measure_0_delta_perc: {
      dot: ".",
      frac: "3",
      int: "33",
      neg: undefined,
      percent: "%",
      suffix: "",
    },
  },
  {
    fruit: "grapes",
    measure_0: 15,
    measure_0_delta: -5,
    measure_0_delta_perc: {
      dot: "",
      frac: "",
      int: "25",
      neg: "-",
      percent: "%",
      suffix: "",
    },
  },
  {
    fruit: "oranges",
    measure_0: 25,
    measure_0_delta: 0,
    measure_0_delta_perc: {
      int: 0,
      neg: "",
      percent: "%",
    },
  },
  {
    fruit: "apple",
    measure_0: 30,
    measure_0_delta: null,
    measure_0_delta_perc: PERC_DIFF.PREV_VALUE_NO_DATA,
  },
  {
    fruit: "guvava",
    measure_0: 35,
    measure_0_delta: 35,
    measure_0_delta_perc: PERC_DIFF.PREV_VALUE_ZERO,
  },
];

describe("computeComparisonValues", () => {
  const values = [
    { fruit: "banana", measure_0: 20 },
    { fruit: "grapes", measure_0: 15 },
    { fruit: "oranges", measure_0: 25 },
    { fruit: "apple", measure_0: 30 },
    { fruit: "guvava", measure_0: 35 },
  ];

  it("should compute comparison values correctly", () => {
    const computedValues = computeComparisonValues(
      comparisonResponse,
      values,
      "fruit",
      "fruit",
      "measure_0"
    );
    expect(computedValues).toEqual(expectedData);
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
