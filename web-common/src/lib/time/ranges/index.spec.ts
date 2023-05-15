import { getAdjustedFetchTime } from "../ranges";
import { V1TimeGrain } from "../../../runtime-client";

const getAdjustedFetchTimeTestCases = [
  {
    test: "should return ",
    start: new Date("2020-01-04T00:00:00.000Z"),
    end: new Date("2020-01-06T00:00:00.000Z"),
    interval: V1TimeGrain.TIME_GRAIN_DAY,
    expected: [
      {
        start: new Date("2020-01-03T00:00:00.000Z"),
        end: new Date("2020-01-07T00:00:00.000Z"),
      },
    ],
  },
];

describe("getAdjustedFetchTime", () => {
  getAdjustedFetchTimeTestCases.forEach((testCase) => {
    it(testCase.test, () => {
      const defaultTimeGrain = getAdjustedFetchTime(
        testCase.start,
        testCase.end,
        testCase.interval
      );
      expect(defaultTimeGrain).toEqual(testCase.expected);
    });
  });
});
