import { getAbbreviationForIANA } from "./index";

const getAbbreviationForIANATestCases = [
  {
    test: "returns the abbreviation for the Indian timezone",
    now: new Date("2022-01-01T00:00:00.000Z"),
    iana: "Asia/Kolkata",
    expected: "IST",
  },
  {
    test: "returns the abbreviation for the Pacific timezone",
    now: new Date("2022-01-01T00:00:00.000Z"),
    iana: "America/Los_Angeles",
    expected: "PST",
  },
  {
    test: "returns the abbreviation accounting for Daylight Savings Time",
    now: new Date("2022-06-01T00:00:00.000Z"),
    iana: "America/Los_Angeles",
    expected: "PDT",
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
