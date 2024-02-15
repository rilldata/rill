import {
  humaniseISODuration,
  isoDurationToTimeRange,
} from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { describe, it, expect } from "vitest";

describe("isoDurationToTimeRange", () => {
  it("hour range, anchored to hour", () => {
    assertDuration(
      "PT4H",
      "2023-10-05T12:00:00.000Z",
      "2023-10-05T09:00:00.000Z",
      "2023-10-05T13:00:00.000Z",
    );
  });

  it("hour range, not anchored", () => {
    assertDuration(
      "PT4H",
      "2023-10-05T12:12:34.000Z",
      "2023-10-05T09:00:00.000Z",
      "2023-10-05T13:00:00.000Z",
    );
  });

  it("day range, anchored to day", () => {
    assertDuration(
      "P14D",
      "2023-10-05T00:00:00.000Z",
      "2023-09-22T00:00:00.000Z",
      "2023-10-06T00:00:00.000Z",
    );
  });

  it("day range, not anchored", () => {
    assertDuration(
      "P14D",
      "2023-10-05T12:12:34.000Z",
      "2023-09-22T00:00:00.000Z",
      "2023-10-06T00:00:00.000Z",
    );
  });

  it("week range, not anchored", () => {
    assertDuration(
      "P4W",
      "2023-10-05T12:12:34.000Z",
      "2023-09-11T00:00:00.000Z",
      "2023-10-09T00:00:00.000Z",
    );
  });
});

describe("humaniseISODuration", () => {
  (
    [
      ["PT4H", "4 Hours"],
      ["P1D", "1 Day"],
      ["P14D", "14 Days"],
      ["P2W", "2 Weeks"],
      ["P1Y3WT6H", "1 Year, 3 Weeks and 6 Hours"],
    ] as Array<[iso: string, humanised: string]>
  ).forEach(([iso, humanised]) => {
    it(`${iso} => ${humanised}`, () => {
      expect(humaniseISODuration(iso)).toBe(humanised);
    });
  });
});

function assertDuration(
  isoDuration: string,
  anchorDate: string,
  expectedStart: string,
  expectedEnd: string,
) {
  const { startTime, endTime } = isoDurationToTimeRange(
    isoDuration,
    new Date(anchorDate),
  );
  expect(startTime.toISOString()).toBe(expectedStart);
  expect(endTime.toISOString()).toBe(expectedEnd);
}
