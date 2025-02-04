import { makeHref } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { describe, it, expect } from "vitest";

describe("makeHref", () => {
  const TestCases: [
    string | number | boolean | null,
    string,
    string | undefined,
  ][] = [
    [null, "http://localhost/dim", undefined],
    ["http://localhost", "http://localhost/dim", "http://localhost"],
    [true, "http://localhost/dim", "http://localhost/dim"],
    [1, "http://localhost/dim", "http://localhost/dim"],
  ];
  for (const [uri, dimValue, expected] of TestCases) {
    it(`makeHref(${uri}, ${dimValue})=${expected}`, () => {
      expect(makeHref(uri as string | boolean | null, dimValue)).toEqual(
        expected,
      );
    });
  }
});
