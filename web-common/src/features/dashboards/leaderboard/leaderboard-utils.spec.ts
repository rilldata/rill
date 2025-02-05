import { makeHref } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { describe, it, expect } from "vitest";

describe("makeHref", () => {
  const TestCases: {
    uri: string | number | boolean | null;
    dimValue: string;
    expected: string | undefined;
  }[] = [
    {
      uri: null,
      dimValue: "http://localhost/dim",
      expected: undefined,
    },
    {
      uri: "http://localhost",
      dimValue: "http://localhost/dim",
      expected: "http://localhost",
    },
    {
      uri: true,
      dimValue: "http://localhost/dim",
      expected: "http://localhost/dim",
    },
    {
      uri: 1,
      dimValue: "http://localhost/dim",
      expected: "http://localhost/dim",
    },
  ];
  for (const { uri, dimValue, expected } of TestCases) {
    it(`makeHref(${uri}, ${dimValue})=${expected}`, () => {
      expect(makeHref(uri as string | boolean | null, dimValue)).toEqual(
        expected,
      );
    });
  }
});
