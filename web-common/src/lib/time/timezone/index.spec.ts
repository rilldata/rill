import { getAbbreviationForIANA, addZoneOffset } from "./index";

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

describe("addZoneOffset", () => {
  it("adds the correct offset for the Indian timezone", () => {
    const dt = new Date("2022-01-01T00:00:00.000Z");
    const iana = "Asia/Kolkata";
    const expected = new Date("2022-01-01T05:30:00.000Z");
    const result = addZoneOffset(dt, iana);
    expect(result).toEqual(expected);
  });

  it("adds the correct offset for the Pacific timezone", () => {
    const dt = new Date("2022-01-01T00:00:00.000Z");
    const iana = "America/Los_Angeles";
    const expected = new Date("2021-12-31T08:00:00.000Z");
    const result = addZoneOffset(dt, iana);
    expect(result).toEqual(expected);
  });

  it("adds the correct offset accounting for Daylight Savings Time", () => {
    const dt = new Date("2022-06-01T00:00:00.000Z");
    const iana = "America/Los_Angeles";
    const expected = new Date("2022-05-31T07:00:00.000Z");
    const result = addZoneOffset(dt, iana);
    expect(result).toEqual(expected);
  });
});
