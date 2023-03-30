import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  computeComparisonValues,
  customSortMeasures,
  getFilterForComparsion,
  updateFilterOnSearch,
} from "./dimension-table-utils";

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
      include: [{ name: "fruit", like: ["%apple%"] }],
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

const measures = [
  "measure_1",
  "measure_10",
  "measure_1_delta",
  "measure_1_delta_perc",
  "measure_2",
  "measure_0",
  "measure_20",
];
const expectedMeasures = [
  "measure_0",
  "measure_1",
  "measure_1_delta",
  "measure_1_delta_perc",
  "measure_2",
  "measure_10",
  "measure_20",
];

describe("customSortMeasures", () => {
  it("should sort the measures in the correct order", () => {
    const sortedMeasures = measures.sort(customSortMeasures);
    expect(sortedMeasures).toEqual({ expectedMeasures });
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

const values = [
  { fruit: "banana", measure_0: 20 },
  { fruit: "grapes", measure_0: 15 },
  { fruit: "oranges", measure_0: 25 },
  { fruit: "apple", measure_0: 30 },
  { fruit: "guvava", measure_0: 35 },
];

const expectedData = [
  {
    fruit: "banana",
    measure_0: 20,
    measure_0_delta: 5,
    measure_0_delta_perc: 33.33,
  },
  {
    fruit: "grapes",
    measure_0: 15,
    measure_0_delta: -5,
    measure_0_delta_perc: -25,
  },
  {
    fruit: "oranges",
    measure_0: 25,
    measure_0_delta: 0,
    measure_0_delta_perc: 0,
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
    measure_0_delta: null,
    measure_0_delta_perc: PERC_DIFF.PREV_VALUE_ZERO,
  },
];

describe("computeComparisonValues", () => {
  it("should compute comparison values correctly", () => {
    const computedValues = computeComparisonValues(comparisonResponse, values);
    expect(computedValues).toEqual(expectedData);
  });
});
