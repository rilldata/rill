import { describe, it, expect } from "vitest";
import { getAbbreviationForIANA, addZoneOffset } from "./index";
import { DateTime } from "luxon";

const getAbbreviationForIANATestCases = [
  {
    test: "returns the abbreviation for the Indian timezone",
    now: DateTime.fromISO("2022-01-01T00:00:00.000Z"),
    iana: "Asia/Kolkata",
    expected: "IST",
  },
  {
    test: "returns the abbreviation for the Pacific timezone",
    now: DateTime.fromISO("2022-01-01T00:00:00.000Z"),
    iana: "America/Los_Angeles",
    expected: "PST",
  },
  {
    test: "returns the abbreviation accounting for Daylight Savings Time",
    now: DateTime.fromISO("2022-06-01T00:00:00.000Z"),
    iana: "America/Los_Angeles",
    expected: "PDT",
  },
];

const addZoneOffsetTestCases = [
  {
    test: "adds the correct offset for the Indian timezone",
    dt: new Date("2022-01-01T00:00:00.000Z"),
    iana: "Asia/Kolkata",
    expected: new Date("2022-01-01T05:30:00.000Z"),
  },
  {
    test: "adds the correct offset for the Pacific timezone",
    dt: new Date("2022-01-01T00:00:00.000Z"),
    iana: "America/Los_Angeles",
    expected: new Date("2021-12-31T16:00:00.000Z"),
  },
  {
    test: "adds the correct offset accounting for Daylight Savings Time",
    dt: new Date("2022-06-01T00:00:00.000Z"),
    iana: "America/Los_Angeles",
    expected: new Date("2022-05-31T17:00:00.000Z"),
  },
];

describe("getAbbreviationForIANA", () => {
  getAbbreviationForIANATestCases.forEach((testCase) => {
    it(testCase.test, () => {
      const abbreviation = getAbbreviationForIANA(testCase.now, testCase.iana);
      expect(abbreviation).toEqual(testCase.expected);
    });
  });
});

describe("addZoneOffset", () => {
  addZoneOffsetTestCases.forEach((testCase) => {
    it(testCase.test, () => {
      const offset = addZoneOffset(testCase.dt, testCase.iana);
      expect(offset).toEqual(testCase.expected);
    });
  });
});
