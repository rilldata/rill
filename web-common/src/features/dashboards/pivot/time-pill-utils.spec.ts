import {
  handleTimeChipClick,
  handleTimeChipDrop,
} from "@rilldata/web-common/features/dashboards/pivot/time-pill-utils";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const DAY = V1TimeGrain.TIME_GRAIN_DAY;
const WEEK = V1TimeGrain.TIME_GRAIN_WEEK;
const MONTH = V1TimeGrain.TIME_GRAIN_MONTH;

const timeChip = (grain: V1TimeGrain) => ({
  id: grain,
  title: TIME_GRAIN[grain].label,
  type: PivotChipType.Time,
});

describe("handleTimeChipClick", () => {
  it("resolves the largest available grain", () => {
    expect(handleTimeChipClick([DAY, MONTH, WEEK])).toEqual(timeChip(MONTH));
  });

  it("resolves nothing when no grain is available", () => {
    // The sidebar's time chip is a placeholder whose id is not a grain, so it
    // must not be added as-is.
    expect(handleTimeChipClick([])).toBeUndefined();
  });
});

describe("handleTimeChipDrop", () => {
  it("resolves the largest available grain when the zone has no time chip", () => {
    expect(handleTimeChipDrop(0, [], [DAY, WEEK])).toEqual(timeChip(WEEK));
  });

  it("resolves the next smaller grain after the time chip to the left", () => {
    expect(handleTimeChipDrop(1, [timeChip(MONTH)], [DAY, WEEK])).toEqual(
      timeChip(WEEK),
    );
  });

  it("resolves nothing when no grain is available", () => {
    expect(handleTimeChipDrop(0, [], [])).toBeUndefined();
    expect(handleTimeChipDrop(1, [timeChip(MONTH)], [])).toBeUndefined();
  });
});
