import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  deriveInterval,
  INHERIT_TIME_RANGE_ALIAS,
} from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import {
  DEFAULT_TIME_RANGE,
  RESOLVED_RILL_TIMES,
  TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/time-controls/test/rill-time-mocks";
import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  createInEffectRoot,
  renderInRuntimeContext,
  useMetricsViewMocks,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import { asyncWait, waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime } from "luxon";
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.stubEnv("TZ", "UTC");

// Wrapped so that a test can tell whether a range was resolved again.
vi.mock(
  "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts",
  async (importOriginal) => {
    const actual =
      await importOriginal<
        typeof import("@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts")
      >();
    return { ...actual, deriveInterval: vi.fn(actual.deriveInterval) };
  },
);

// ---------------------------------------------------------------------------
// Test metrics views
//
// AdBids has a time dimension and resolves rilltime expressions to the intervals in
// `RESOLVED_RILL_TIMES`. Its mirror resolves the default range to a window ending a day later,
// which makes the choice of interval across metrics views observable.
// The third metrics view has no time dimension at all.
// ---------------------------------------------------------------------------

const AD_BIDS_MIRROR_METRICS_NAME = "AdBids_mirror_metrics";
const AD_BIDS_NO_TIME_METRICS_NAME = "AdBids_no_time_metrics";

const mocks = useMetricsViewMocks({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT_WITH_TIME,
  [AD_BIDS_MIRROR_METRICS_NAME]: AD_BIDS_METRICS_INIT_WITH_TIME,
  [AD_BIDS_NO_TIME_METRICS_NAME]: AD_BIDS_METRICS_INIT,
});
mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, TIME_RANGE_SUMMARY);
mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
  ...RESOLVED_RILL_TIMES,
  // The default picked from the time range summary when the yaml has none:
  // the summary spans about 90 days, which lands on quarter to date.
  [TimeRangePreset.QUARTER_TO_DATE]: {
    start: "2024-01-01T00:00:00.000Z",
    end: "2024-04-01T00:00:00.000Z",
    grain: V1TimeGrain.TIME_GRAIN_DAY,
  },
});
mocks.mockTimeRangeSummary(AD_BIDS_MIRROR_METRICS_NAME, TIME_RANGE_SUMMARY);
mocks.mockResolvedRillTimes(AD_BIDS_MIRROR_METRICS_NAME, {
  [DEFAULT_TIME_RANGE]: {
    start: "2024-03-26T00:00:00.000Z",
    end: "2024-04-02T00:00:00.000Z",
    grain: V1TimeGrain.TIME_GRAIN_DAY,
  },
});

// Each test gets its own provider and manager.
// The queries behind the provider are cached globally, so only the first provider waits on fetches.
let metricsViewsProvider: MetricsViewsProvider;
const cleanups: (() => void)[] = [];

beforeEach(async () => {
  metricsViewsProvider = await createReadyMetricsViewsProvider([
    AD_BIDS_METRICS_NAME,
  ]);
});

afterEach(() => {
  cleanups.splice(0).forEach((destroy) => destroy());
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type CreateOptions = {
  provider?: MetricsViewsProvider;
  yamlConfigProvider?: YAMLConfigProvider;
  config?: TimeFiltersConfig;
  parent?: TimeFilterManager;
};

/**
 * A provider over `metricsViewNames`, torn down after the test.
 * Resolves once the specs and the time range summaries have landed.
 */
async function createReadyMetricsViewsProvider(metricsViewNames: string[]) {
  // MetricsViewsProvider calls createQuery, so it has to be created inside a component.
  const rendered = renderInRuntimeContext(
    ({ runtimeClient }) =>
      new MetricsViewsProvider(runtimeClient, metricsViewNames),
  );
  cleanups.push(() => {
    rendered.value.cleanup();
    rendered.destroy();
  });
  expect(await waitUntil(() => rendered.value.ready, 2000, 5)).toBe(true);
  return rendered.value;
}

/** A yaml config with the default time range the dashboards in `rill-time-mocks.ts` start on. */
function yamlWithDefaultTimeRange() {
  const yamlConfigProvider = new YAMLConfigProvider();
  yamlConfigProvider.defaultTimeRange = DEFAULT_TIME_RANGE;
  return yamlConfigProvider;
}

/** A real time filter manager over a ready provider, torn down after the test. */
function createTimeFilterManager({
  provider = metricsViewsProvider,
  yamlConfigProvider = yamlWithDefaultTimeRange(),
  config = {},
  parent,
}: CreateOptions = {}) {
  const { value, destroy } = createInEffectRoot(
    () =>
      new TimeFilterManager(
        provider.runtimeClient,
        provider,
        yamlConfigProvider,
        config,
        parent,
      ),
  );
  cleanups.push(destroy);
  return value;
}

async function waitForReady(manager: TimeFilterManager) {
  expect(await waitUntil(() => manager.ready, 2000, 5)).toBe(true);
}

/** Resolves once `timeRange` has been resolved by the runtime and applied. */
async function waitForTimeRange(
  manager: TimeFilterManager,
  timeRange: string | undefined,
) {
  const applied = await waitUntil(
    () => manager.timeRange === timeRange && !!manager.interval,
    2000,
    5,
  );
  expect(applied, `time range to be "${timeRange}"`).toBe(true);
}

/** Records every param set the manager reports to the url. */
function recordChanges(manager: TimeFilterManager) {
  const changes: Record<string, string>[] = [];
  cleanups.push(
    manager.storeSync.on("change", (params) => {
      changes.push(Object.fromEntries(params));
    }),
  );
  return changes;
}

/** `stateChanged` reports on the next microtask. */
function flushStateChanges() {
  return asyncWait(0);
}

function paramsOf(manager: TimeFilterManager) {
  const params = new URLSearchParams();
  manager.applyFilterToParams(params);
  return Object.fromEntries(params);
}

function isoOf(dateTime: DateTime | undefined) {
  return dateTime?.toJSDate().toISOString();
}

function intervalOf(manager: TimeFilterManager) {
  return {
    start: isoOf(manager.interval?.start),
    end: isoOf(manager.interval?.end),
  };
}

function comparisonIntervalOf(manager: TimeFilterManager) {
  return {
    start: isoOf(manager.comparisonInterval?.start),
    end: isoOf(manager.comparisonInterval?.end),
  };
}

function resolvedIntervalOf(timeRange: string) {
  const { start, end } = RESOLVED_RILL_TIMES[timeRange];
  return { start, end };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("setUrlParams", () => {
  it("applies the yaml default time range when the url has none", async () => {
    const manager = createTimeFilterManager();

    // The tracker skips params equal to the ones it holds, and it starts out empty,
    // so an empty url goes to the manager directly.
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(manager.urlTimeRange).toBe(DEFAULT_TIME_RANGE);
    expect(intervalOf(manager)).toEqual(resolvedIntervalOf(DEFAULT_TIME_RANGE));
    // The grain comes from the snap of the `as of` clause.
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    expect(manager.apiTimeRange).toEqual({
      ...resolvedIntervalOf(DEFAULT_TIME_RANGE),
      timeZone: "UTC",
      timeDimension: undefined,
    });
  });

  it("resolves the time range in the url", async () => {
    const manager = createTimeFilterManager();

    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: "4W as of latest/D+1D" }),
    );
    await waitForTimeRange(manager, "4W as of latest/D+1D");

    expect(intervalOf(manager)).toEqual(
      resolvedIntervalOf("4W as of latest/D+1D"),
    );
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_DAY);
  });

  it("reads the grain, time zone and time dimension from the url", async () => {
    const manager = createTimeFilterManager();

    manager.storeSync.setUrlParams(
      new URLSearchParams({
        tr: "4W as of latest/D+1D",
        grain: "week",
        tz: "America/New_York",
        td: AD_BIDS_TIMESTAMP_DIMENSION,
      }),
    );
    await waitForTimeRange(manager, "4W as of latest/D+1D");

    // Week is a grain the 4 week interval allows, so the url grain wins over the snap.
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_WEEK);
    expect(manager.timeZone).toBe("America/New_York");
    expect(manager.interval?.start.zoneName).toBe("America/New_York");
    expect(manager.apiTimeRange).toEqual({
      ...resolvedIntervalOf("4W as of latest/D+1D"),
      timeZone: "America/New_York",
      timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
    });
  });

  it("falls back to the snap grain when the url grain is not allowed for the interval", async () => {
    const manager = createTimeFilterManager();

    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: "12h as of latest/h+1h", grain: "month" }),
    );
    await waitForTimeRange(manager, "12h as of latest/h+1h");

    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
  });

  it("falls back to the yaml default time zone", async () => {
    const yamlConfigProvider = yamlWithDefaultTimeRange();
    yamlConfigProvider.defaultTimeZone = "Asia/Kathmandu";
    const manager = createTimeFilterManager({ yamlConfigProvider });

    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(manager.timeZone).toBe("Asia/Kathmandu");
    expect(manager.apiTimeRange.timeZone).toBe("Asia/Kathmandu");
  });

  it("computes the contiguous comparison even when comparison is off", async () => {
    const manager = createTimeFilterManager();

    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(manager.showComparison).toBe(false);
    expect(manager.comparisonTimeRange).toBe(TimeComparisonOption.CONTIGUOUS);
    // The 7 days right before the selected ones.
    expect(comparisonIntervalOf(manager)).toEqual({
      start: "2024-03-18T00:00:00.000Z",
      end: "2024-03-25T00:00:00.000Z",
    });
    expect(manager.apiComparisonTimeRange).toEqual({
      start: "2024-03-18T00:00:00.000Z",
      end: "2024-03-25T00:00:00.000Z",
      timeZone: "UTC",
    });
  });

  it("turns comparison on from the url", async () => {
    const manager = createTimeFilterManager();

    manager.storeSync.setUrlParams(
      new URLSearchParams({
        tr: DEFAULT_TIME_RANGE,
        compare_tr: TimeComparisonOption.WEEK,
      }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(manager.showComparison).toBe(true);
    expect(manager.comparisonTimeRange).toBe(TimeComparisonOption.WEEK);
    expect(comparisonIntervalOf(manager)).toEqual({
      start: "2024-03-18T00:00:00.000Z",
      end: "2024-03-25T00:00:00.000Z",
    });
  });

  it("leaves the time range empty with skipDefaultTimeRange", async () => {
    const manager = createTimeFilterManager({
      config: { skipDefaultTimeRange: true },
    });

    manager.setUrlParams(new URLSearchParams());
    await asyncWait(20);

    expect(manager.urlTimeRange).toBeUndefined();
    expect(manager.timeRange).toBeUndefined();
    expect(manager.interval).toBeUndefined();
    expect(manager.apiTimeRange.start).toBeUndefined();
    expect(manager.apiTimeRange.end).toBeUndefined();
  });

  it("clears a previous time range once the url drops it with skipDefaultTimeRange", async () => {
    const manager = createTimeFilterManager({
      config: { skipDefaultTimeRange: true },
    });

    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    manager.storeSync.setUrlParams(new URLSearchParams());

    expect(manager.timeRange).toBeUndefined();
    expect(manager.interval).toBeUndefined();
  });

  it("reads the highlighted range from the url", async () => {
    const manager = createTimeFilterManager();

    manager.storeSync.setUrlParams(
      new URLSearchParams({
        tr: DEFAULT_TIME_RANGE,
        highlighted_tr: "2024-03-26T00:00:00.000Z,2024-03-27T00:00:00.000Z",
      }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    // Applying the time range clears the scrub, which must not drop the one in the url.
    expect(manager.timeStart).toBe("2024-03-26T00:00:00.000Z");
    expect(manager.timeEnd).toBe("2024-03-27T00:00:00.000Z");
    expect(manager.scrubInterval?.isScrubbing).toBe(false);
    // The selected range itself is untouched.
    expect(intervalOf(manager)).toEqual(resolvedIntervalOf(DEFAULT_TIME_RANGE));
    expect(paramsOf(manager).highlighted_tr).toBe(
      "2024-03-26T00:00:00.000Z,2024-03-27T00:00:00.000Z",
    );
  });

  it("clears the highlighted range once the url drops it", async () => {
    const manager = createTimeFilterManager({
      config: { skipDefaultTimeRange: true },
    });
    manager.storeSync.setUrlParams(
      new URLSearchParams({
        tr: DEFAULT_TIME_RANGE,
        highlighted_tr: "2024-03-26T00:00:00.000Z,2024-03-27T00:00:00.000Z",
      }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    // No time range is applied here, so nothing else clears the scrub.
    manager.storeSync.setUrlParams(new URLSearchParams({ grain: "day" }));

    expect(manager.scrubInterval).toBeUndefined();
    expect(manager.lastDefinedScrubInterval).toBeUndefined();
  });

  it("ignores the url grain with skipTimeGrain", async () => {
    const manager = createTimeFilterManager({
      config: { skipTimeGrain: true },
    });

    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: "4W as of latest/D+1D", grain: "week" }),
    );
    await waitForTimeRange(manager, "4W as of latest/D+1D");

    // The resolved grain is still tracked, it is just not read from or written to the url.
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    expect(paramsOf(manager)).not.toHaveProperty(
      ExploreStateURLParams.TimeGrain,
    );
  });
});

describe("delayed load", () => {
  /**
   * A manager created before its provider has any data, as on a dashboard's first load.
   * `beforeSpecsLoad` runs right after the manager is created.
   */
  function createBeforeLoad(
    yamlConfigProvider: YAMLConfigProvider,
    beforeSpecsLoad: (manager: TimeFilterManager) => void,
  ) {
    // Drop the responses cached by the `beforeEach` provider so that this one has to fetch.
    queryClient.clear();
    const rendered = renderInRuntimeContext(({ runtimeClient }) => {
      const provider = new MetricsViewsProvider(runtimeClient, [
        AD_BIDS_METRICS_NAME,
      ]);
      const manager = new TimeFilterManager(
        runtimeClient,
        provider,
        yamlConfigProvider,
      );
      beforeSpecsLoad(manager);
      return { provider, manager };
    });
    cleanups.push(() => {
      rendered.value.provider.cleanup();
      rendered.destroy();
    });
    return rendered.value.manager;
  }

  it("holds the url params until the time range summary lands", async () => {
    const manager = createBeforeLoad(yamlWithDefaultTimeRange(), (manager) =>
      manager.storeSync.setUrlParams(
        new URLSearchParams({ tr: "4W as of latest/D+1D", grain: "week" }),
      ),
    );
    expect(manager.ready).toBe(false);
    expect(manager.urlTimeRange).toBeUndefined();

    await waitForReady(manager);
    await waitForTimeRange(manager, "4W as of latest/D+1D");
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_WEEK);
  });

  it("picks the default time range from the time range summary once it lands", async () => {
    // Without a yaml default, the default depends on the span of the data,
    // so it cannot be picked before the summary is loaded.
    const manager = createBeforeLoad(new YAMLConfigProvider(), (manager) =>
      manager.storeSync.setUrlParams(
        new URLSearchParams({ tz: "America/New_York" }),
      ),
    );
    expect(manager.ready).toBe(false);

    await waitForReady(manager);
    await waitForTimeRange(manager, TimeRangePreset.QUARTER_TO_DATE);
    expect(manager.timeZone).toBe("America/New_York");
    expect(intervalOf(manager)).toEqual({
      start: "2024-01-01T00:00:00.000Z",
      end: "2024-04-01T00:00:00.000Z",
    });
  });
});

describe("applyFilterToParams", () => {
  it("writes the time params back", async () => {
    const manager = createTimeFilterManager();
    const params = {
      tr: "4W as of latest/D+1D",
      grain: "week",
      tz: "America/New_York",
      td: AD_BIDS_TIMESTAMP_DIMENSION,
      compare_tr: TimeComparisonOption.WEEK,
    };

    manager.storeSync.setUrlParams(new URLSearchParams(params));
    await waitForTimeRange(manager, "4W as of latest/D+1D");

    expect(paramsOf(manager)).toEqual(params);
  });

  it("omits the default time zone and a comparison that is off", async () => {
    const manager = createTimeFilterManager();

    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE, tz: "UTC" }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    // The comparison range is still set internally, it is just not shown.
    expect(manager.comparisonTimeRange).toBe(TimeComparisonOption.CONTIGUOUS);
    expect(paramsOf(manager)).toEqual({
      tr: DEFAULT_TIME_RANGE,
      grain: "day",
    });
  });

  it("removes params the manager no longer holds", () => {
    const manager = createTimeFilterManager({
      config: { skipDefaultTimeRange: true },
    });
    manager.setUrlParams(new URLSearchParams());

    const searchParams = new URLSearchParams({
      tr: DEFAULT_TIME_RANGE,
      tz: "America/New_York",
      td: AD_BIDS_TIMESTAMP_DIMENSION,
      compare_tr: TimeComparisonOption.WEEK,
      highlighted_tr: "2024-03-26T00:00:00.000Z,2024-03-27T00:00:00.000Z",
      // Not a time param, so it is left alone.
      view: "tdd",
    });
    manager.applyFilterToParams(searchParams);

    expect(Object.fromEntries(searchParams)).toEqual({ view: "tdd" });
  });
});

describe("onSelectRange", () => {
  it("keeps the snap of the current range for a coarser range", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const changes = recordChanges(manager);

    await manager.onSelectRange("4W");
    await flushStateChanges();

    expect(manager.timeRange).toBe("4W as of latest/D+1D");
    expect(intervalOf(manager)).toEqual(
      resolvedIntervalOf("4W as of latest/D+1D"),
    );
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    // One url update per selection.
    expect(changes).toEqual([{ tr: "4W as of latest/D+1D", grain: "day" }]);
  });

  it("narrows the snap to the grain of a finer range", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    await manager.onSelectRange("12h");

    expect(manager.timeRange).toBe("12h as of latest/h+1h");
    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
  });

  it("recomputes the comparison for the new range", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({
        tr: DEFAULT_TIME_RANGE,
        compare_tr: TimeComparisonOption.WEEK,
      }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    await manager.onSelectRange("24h");

    expect(manager.timeRange).toBe("24h as of latest/h+1h");
    expect(manager.comparisonTimeRange).toBe(TimeComparisonOption.WEEK);
    expect(comparisonIntervalOf(manager)).toEqual({
      start: "2024-03-23T15:00:00.000Z",
      end: "2024-03-24T15:00:00.000Z",
    });
  });

  it("clears the scrubbed range", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    manager.onScrubRange({
      start: DateTime.fromISO("2024-03-26T00:00:00.000Z"),
      end: DateTime.fromISO("2024-03-27T00:00:00.000Z"),
      isScrubbing: false,
    });

    await manager.onSelectRange("4W");

    expect(manager.scrubInterval).toBeUndefined();
    expect(manager.lastDefinedScrubInterval).toBeUndefined();
  });
});

describe("onSelectAsOfOption", () => {
  it("moves the snap offset to the start of the grain", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    await manager.onSelectAsOfOption(manager.ref, false);

    expect(manager.timeRange).toBe("7D as of latest/D");
    expect(intervalOf(manager)).toEqual(
      resolvedIntervalOf("7D as of latest/D"),
    );
    expect(manager.snapToEnd).toBe(false);
  });
});

describe("parsed time range", () => {
  it("exposes the reference, snap and truncation grain of the range", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(manager.ref).toBe("latest");
    expect(manager.snapToEnd).toBe(true);
    expect(manager.truncationGrain).toBe(V1TimeGrain.TIME_GRAIN_DAY);
  });
});

describe("onSelectGrain", () => {
  it("sets the grain and reports it to the url", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: "4W as of latest/D+1D" }),
    );
    await waitForTimeRange(manager, "4W as of latest/D+1D");
    const changes = recordChanges(manager);

    manager.onSelectGrain(V1TimeGrain.TIME_GRAIN_WEEK);
    await flushStateChanges();

    expect(manager.timeGrain).toBe(V1TimeGrain.TIME_GRAIN_WEEK);
    expect(changes).toEqual([{ tr: "4W as of latest/D+1D", grain: "week" }]);
  });
});

describe("onSelectTimeDimension", () => {
  it("sets the time dimension and reports it to the url", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const changes = recordChanges(manager);

    manager.onSelectTimeDimension(AD_BIDS_TIMESTAMP_DIMENSION);
    await flushStateChanges();

    expect(manager.apiTimeRange.timeDimension).toBe(
      AD_BIDS_TIMESTAMP_DIMENSION,
    );
    expect(changes).toEqual([
      { tr: DEFAULT_TIME_RANGE, grain: "day", td: AD_BIDS_TIMESTAMP_DIMENSION },
    ]);
  });

  it("resolves the range again against the new time dimension", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    vi.mocked(deriveInterval).mockClear();

    manager.onSelectTimeDimension(AD_BIDS_TIMESTAMP_DIMENSION);
    await asyncWait(20);

    expect(deriveInterval).toHaveBeenCalledWith(
      DEFAULT_TIME_RANGE,
      expect.anything(),
      AD_BIDS_METRICS_NAME,
      "UTC",
      AD_BIDS_TIMESTAMP_DIMENSION,
      undefined,
    );
  });
});

describe("onSelectZone", () => {
  it("sets the time zone and reports it to the url", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const changes = recordChanges(manager);

    manager.onSelectZone("America/New_York");
    await flushStateChanges();

    expect(manager.timeZone).toBe("America/New_York");
    expect(manager.apiTimeRange.timeZone).toBe("America/New_York");
    expect(changes).toEqual([
      { tr: DEFAULT_TIME_RANGE, grain: "day", tz: "America/New_York" },
    ]);
  });

  it("resolves the range again in the new time zone", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    manager.onSelectZone("America/New_York");
    await asyncWait(20);

    expect(manager.interval?.start.zoneName).toBe("America/New_York");
  });
});

describe("pan", () => {
  it("allows panning both ways for a range inside the data", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: "7D as of latest/D" }),
    );
    await waitForTimeRange(manager, "7D as of latest/D");

    // 2024-03-24 to 2024-03-31, inside 2024-01-01 to 2024-03-31T14:30.
    expect(manager.canPanLeft).toBe(true);
    expect(manager.canPanRight).toBe(true);
  });

  it("does not allow panning right past the end of the data", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    // Padded to the end of the day, so it ends after the data does.
    expect(manager.canPanLeft).toBe(true);
    expect(manager.canPanRight).toBe(false);
  });
});

describe("comparison", () => {
  it("onSelectComparisonRange turns comparison on with the selected range", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const changes = recordChanges(manager);

    manager.onSelectComparisonRange(TimeComparisonOption.WEEK);
    await flushStateChanges();

    expect(manager.showComparison).toBe(true);
    expect(manager.comparisonTimeRange).toBe(TimeComparisonOption.WEEK);
    expect(comparisonIntervalOf(manager)).toEqual({
      start: "2024-03-18T00:00:00.000Z",
      end: "2024-03-25T00:00:00.000Z",
    });
    expect(changes).toEqual([
      {
        tr: DEFAULT_TIME_RANGE,
        grain: "day",
        compare_tr: TimeComparisonOption.WEEK,
      },
    ]);
  });

  it("onToggleShowComparison and setShowComparison flip the comparison", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const changes = recordChanges(manager);

    manager.onToggleShowComparison();
    await flushStateChanges();
    expect(manager.showComparison).toBe(true);

    manager.setShowComparison(false);
    await flushStateChanges();
    expect(manager.showComparison).toBe(false);

    expect(changes).toEqual([
      {
        tr: DEFAULT_TIME_RANGE,
        grain: "day",
        compare_tr: TimeComparisonOption.CONTIGUOUS,
      },
      { tr: DEFAULT_TIME_RANGE, grain: "day" },
    ]);
  });

  it("switches to the contiguous comparison for an absolute time range", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    // Absolute ranges are not resolved by the runtime mock, so set the range directly.
    manager.timeRange = "2024-03-01T00:00:00Z to 2024-03-08T00:00:00Z";

    manager.onSelectComparisonRange(TimeComparisonOption.WEEK);

    expect(manager.comparisonTimeRange).toBe(TimeComparisonOption.CONTIGUOUS);
  });

  it("offers the comparisons that fit in the data", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    // For 7 days, the previous period is the previous week, so only the latter is listed.
    // The previous day is shorter than the range,
    // and the previous quarter and year start before the data does on 2024-01-01.
    expect(manager.comparisonTimeRangeOptions.map(({ name }) => name)).toEqual([
      TimeComparisonOption.WEEK,
      TimeComparisonOption.MONTH,
    ]);
  });

  it("offers only the yaml comparisons for a yaml time range", async () => {
    const yamlConfigProvider = yamlWithDefaultTimeRange();
    yamlConfigProvider.timeRanges = [
      {
        range: DEFAULT_TIME_RANGE,
        comparisonTimeRanges: [
          { offset: TimeComparisonOption.WEEK },
          { offset: TimeComparisonOption.DAY },
        ],
      },
    ];
    const manager = createTimeFilterManager({ yamlConfigProvider });
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(
      manager.comparisonTimeRangeOptions.map(({ name }) => name),
      // The previous day is still dropped, since it is shorter than the range.
    ).toEqual([TimeComparisonOption.WEEK]);
  });

  it("offers the custom comparison with allowCustomTimeRange", async () => {
    const manager = createTimeFilterManager({
      config: { allowCustomTimeRange: true },
    });
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(
      manager.comparisonTimeRangeOptions.map(({ name }) => name),
    ).toContain(TimeComparisonOption.CUSTOM);
  });
});

describe("scrub", () => {
  it("records the range only once scrubbing ends", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const changes = recordChanges(manager);
    const start = DateTime.fromISO("2024-03-26T00:00:00.000Z");
    const end = DateTime.fromISO("2024-03-27T00:00:00.000Z");

    manager.onScrubRange({ start, end, isScrubbing: true });
    await flushStateChanges();

    expect(manager.scrubInterval).toEqual({ start, end, isScrubbing: true });
    expect(manager.lastDefinedScrubInterval).toBeUndefined();
    expect(changes).toEqual([]);

    manager.onScrubRange({ start, end, isScrubbing: false });
    await flushStateChanges();

    expect(isoOf(manager.lastDefinedScrubInterval?.start)).toBe(
      "2024-03-26T00:00:00.000Z",
    );
    expect(isoOf(manager.lastDefinedScrubInterval?.end)).toBe(
      "2024-03-27T00:00:00.000Z",
    );
    // The scrubbed range narrows the range queries use.
    expect(manager.timeStart).toBe("2024-03-26T00:00:00.000Z");
    expect(manager.timeEnd).toBe("2024-03-27T00:00:00.000Z");
    expect(changes).toEqual([
      {
        tr: DEFAULT_TIME_RANGE,
        grain: "day",
        highlighted_tr: "2024-03-26T00:00:00.000Z,2024-03-27T00:00:00.000Z",
      },
    ]);
  });

  it("orders a range scrubbed right to left", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    manager.onScrubRange({
      start: DateTime.fromISO("2024-03-27T00:00:00.000Z"),
      end: DateTime.fromISO("2024-03-26T00:00:00.000Z"),
      isScrubbing: false,
    });

    expect(manager.timeStart).toBe("2024-03-26T00:00:00.000Z");
    expect(manager.timeEnd).toBe("2024-03-27T00:00:00.000Z");
  });

  it("resetScrubRange goes back to the selected range", async () => {
    const manager = createTimeFilterManager();
    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    manager.onScrubRange({
      start: DateTime.fromISO("2024-03-26T00:00:00.000Z"),
      end: DateTime.fromISO("2024-03-27T00:00:00.000Z"),
      isScrubbing: false,
    });

    manager.resetScrubRange();

    expect(manager.scrubInterval).toBeUndefined();
    expect(manager.lastDefinedScrubInterval).toBeUndefined();
    expect({ start: manager.timeStart, end: manager.timeEnd }).toEqual(
      resolvedIntervalOf(DEFAULT_TIME_RANGE),
    );
  });
});

describe("inherit", () => {
  function createParentAndChild() {
    const parent = createTimeFilterManager();
    const child = createTimeFilterManager({ parent });
    return { parent, child };
  }

  it("resolves to the time range of the parent", async () => {
    const { parent, child } = createParentAndChild();
    parent.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(parent, DEFAULT_TIME_RANGE);

    child.storeSync.setUrlParams(
      new URLSearchParams({ tr: INHERIT_TIME_RANGE_ALIAS }),
    );
    await waitForTimeRange(child, DEFAULT_TIME_RANGE);

    expect(child.urlTimeRange).toBe(INHERIT_TIME_RANGE_ALIAS);
    expect(intervalOf(child)).toEqual(resolvedIntervalOf(DEFAULT_TIME_RANGE));
    // The url keeps the alias, so the child keeps following the parent.
    expect(paramsOf(child).tr).toBe(INHERIT_TIME_RANGE_ALIAS);
  });

  it("follows the parent when its time range changes", async () => {
    const { parent, child } = createParentAndChild();
    parent.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(parent, DEFAULT_TIME_RANGE);
    child.storeSync.setUrlParams(
      new URLSearchParams({ tr: INHERIT_TIME_RANGE_ALIAS }),
    );
    await waitForTimeRange(child, DEFAULT_TIME_RANGE);

    await parent.onSelectRange("12h");
    await flushStateChanges();

    await waitForTimeRange(child, "12h as of latest/h+1h");
    expect(child.urlTimeRange).toBe(INHERIT_TIME_RANGE_ALIAS);
  });

  it("follows the parent to a range with the same grain", async () => {
    const { parent, child } = createParentAndChild();
    parent.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(parent, DEFAULT_TIME_RANGE);
    child.storeSync.setUrlParams(
      new URLSearchParams({ tr: INHERIT_TIME_RANGE_ALIAS }),
    );
    await waitForTimeRange(child, DEFAULT_TIME_RANGE);

    await parent.onSelectRange("4W");
    await flushStateChanges();
    await asyncWait(20);

    expect(child.timeRange).toBe("4W as of latest/D+1D");
  });

  it("stops following the parent once it has its own time range", async () => {
    const { parent, child } = createParentAndChild();
    parent.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(parent, DEFAULT_TIME_RANGE);
    child.storeSync.setUrlParams(
      new URLSearchParams({ tr: "12h as of latest/h+1h" }),
    );
    await waitForTimeRange(child, "12h as of latest/h+1h");

    await parent.onSelectRange("4W");
    await flushStateChanges();
    await asyncWait(20);

    expect(child.timeRange).toBe("12h as of latest/h+1h");
  });
});

describe("multiple metrics views", () => {
  it("uses the interval that ends last", async () => {
    const manager = createTimeFilterManager({
      provider: await createReadyMetricsViewsProvider([
        AD_BIDS_METRICS_NAME,
        AD_BIDS_MIRROR_METRICS_NAME,
      ]),
    });

    manager.setUrlParams(new URLSearchParams());
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);

    expect(intervalOf(manager)).toEqual({
      start: "2024-03-26T00:00:00.000Z",
      end: "2024-04-02T00:00:00.000Z",
    });
  });
});

describe("hasTimeSeries", () => {
  it("is true when a metrics view has a time dimension", async () => {
    const manager = createTimeFilterManager({
      provider: await createReadyMetricsViewsProvider([
        AD_BIDS_NO_TIME_METRICS_NAME,
        AD_BIDS_METRICS_NAME,
      ]),
    });

    expect(manager.hasTimeSeries).toBe(true);
  });

  it("is false when no metrics view has a time dimension", async () => {
    // Metrics views without a time dimension have no summary to wait on,
    // so the provider is ready as soon as the specs land.
    const manager = createTimeFilterManager({
      provider: await createReadyMetricsViewsProvider([
        AD_BIDS_NO_TIME_METRICS_NAME,
      ]),
    });

    expect(manager.hasTimeSeries).toBe(false);
  });
});

describe("getTimeControlStore", () => {
  it("exposes the comparison range only while comparison is shown", async () => {
    const manager = createTimeFilterManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams({ tr: DEFAULT_TIME_RANGE }),
    );
    await waitForTimeRange(manager, DEFAULT_TIME_RANGE);
    const store = manager.getTimeControlStore();

    expect(get(store)).toMatchObject({
      timeRange: DEFAULT_TIME_RANGE,
      timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      timeZone: "UTC",
      showComparison: false,
      apiComparisonTimeRange: undefined,
      hasTimeSeries: true,
      ready: true,
    });

    manager.setShowComparison(true);

    expect(get(store)).toMatchObject({
      showComparison: true,
      apiComparisonTimeRange: {
        start: "2024-03-18T00:00:00.000Z",
        end: "2024-03-25T00:00:00.000Z",
        timeZone: "UTC",
      },
    });
  });
});
