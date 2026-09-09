import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { PageMockForComponentTests } from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import {
  DEFAULT_TIME_RANGE,
  resolvedTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/test/rill-time-mocks";
import { selectTimeRange } from "@rilldata/web-common/features/dashboards/time-controls/test/time-filter-test-utils";
import { describe, expect, it } from "vitest";

type TestTimeRange = {
  name: string;
  start: string;
  end: string;
};

/**
 * How a test reads the filter state back.
 * Explore takes it from the dashboard store the filter bar writes to,
 * while a standalone filter bar has the manager itself as the only holder of the state.
 * TODO: this should not be needed and all variants should read/write from the manager.
 */
export interface TimeFilterManagerAccessor {
  getTimeRange(): TestTimeRange | undefined;
  getTimeGrain(): V1TimeGrain | undefined;
  getComparisonTimeRange(): TestTimeRange | undefined;
}

/**
 * One way of rendering the filter bar.
 * The tests in this file drive the UI identically for every variant,
 * so a variant only covers where the filter state ends up and how it reaches the url.
 */
export interface TimeFiltersVariant {
  /** Renders the filter bar and resolves once it is ready for interaction. */
  render(initUrlSearch?: string): Promise<void>;
  timeFilterManager: TimeFilterManagerAccessor;
  /** Url search of the dashboard before any filter is applied. */
  initialUrlSearch: string;
  /** Url searches the dashboard goes through before the first filter is applied. */
  initialUrlSearchHistory: string[];
  /** Url search once `timeParams`, the time params of the filter bar, are applied. */
  urlSearchWithTimeParams(timeParams: Record<string, string>): string;
  /** A method rather than a field, since the page mock is rebuilt for every test. */
  pageMock(): PageMockForComponentTests;
  /** If true, the url search assertions will be no-op. */
  noUrlSync?: boolean;
}

/**
 * The assertions that differ between variants, bound to one of them.
 * Every test asserts through these, so the test bodies themselves stay variant agnostic.
 */
function variantAssertions(variant: TimeFiltersVariant) {
  return {
    initialUrlSearch: variant.initialUrlSearch,
    urlSearchWithTimeParams: (timeParams: Record<string, string>) =>
      variant.urlSearchWithTimeParams(timeParams),

    assertTimeRange: (expected: TestTimeRange) =>
      expect(variant.timeFilterManager.getTimeRange()).toEqual(expected),

    assertTimeGrain: (expected: V1TimeGrain) =>
      expect(variant.timeFilterManager.getTimeGrain()).toEqual(expected),

    assertComparisonTimeRange: (expected: TestTimeRange) =>
      expect(variant.timeFilterManager.getComparisonTimeRange()).toEqual(
        expected,
      ),

    assertUrlSearch: (expectedSearch: string) => {
      if (variant.noUrlSync) return;
      variant.pageMock().assertSearchParams(expectedSearch);
    },

    /**
     * Asserts that the dashboard went through exactly `searches` after it loaded,
     * which is how the tests catch extra history entries.
     */
    assertUrlSearchHistory: (...searches: string[]) => {
      if (variant.noUrlSync) return;
      expect(variant.pageMock().urlSearchHistory).toEqual([
        ...variant.initialUrlSearchHistory,
        ...searches,
      ]);
    },
  };
}

export function testTimeRangeFilters(variant: TimeFiltersVariant) {
  const {
    urlSearchWithTimeParams,
    assertTimeRange,
    assertTimeGrain,
    assertUrlSearch,
    assertUrlSearchHistory,
  } = variantAssertions(variant);

  describe("Time range filters", () => {
    it("Should change time range and keep snap", async () => {
      await variant.render();

      // The dashboard loads on the yaml default, snapped to the day and padded to include today.
      assertTimeRange(resolvedTimeRange(DEFAULT_TIME_RANGE));
      assertTimeGrain(V1TimeGrain.TIME_GRAIN_DAY);

      // A range coarser than the snap keeps the whole `as of` clause, so the new range is anchored
      // at the same point in time as the one it replaces.
      await selectTimeRange(/Last 4 weeks/);

      const weeksTimeRange = "4W as of latest/D+1D";
      assertTimeRange(resolvedTimeRange(weeksTimeRange));
      // Day is still a grain the wider interval allows, so the grain carries over as well.
      assertTimeGrain(V1TimeGrain.TIME_GRAIN_DAY);
      const weeksUrlSearch = urlSearchWithTimeParams({ tr: weeksTimeRange });
      assertUrlSearch(weeksUrlSearch);
      // Applying the range should add a single entry to history.
      assertUrlSearchHistory(weeksUrlSearch);

      // A range finer than the snap narrows it to the grain of the range, keeping the padding.
      await selectTimeRange(/Last 24 hours/);

      const hoursTimeRange = "24h as of latest/h+1h";
      assertTimeRange(resolvedTimeRange(hoursTimeRange));
      assertTimeGrain(V1TimeGrain.TIME_GRAIN_DAY);
      const hoursUrlSearch = urlSearchWithTimeParams({ tr: hoursTimeRange });
      assertUrlSearch(hoursUrlSearch);
      assertUrlSearchHistory(weeksUrlSearch, hoursUrlSearch);
    });
  });
}
