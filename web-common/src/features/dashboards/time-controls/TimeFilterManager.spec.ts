import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { INHERIT_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import {
  mapTimeManagerComparisonRangeToV1TimeRange,
  mapTimeManagerRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import {
  DEFAULT_TIME_RANGE,
  RESOLVED_RILL_TIMES,
  TIME_RANGE_SUMMARY,
  YAML_TIME_ZONES,
} from "@rilldata/web-common/features/dashboards/time-controls/test/rill-time-mocks";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  renderInRuntimeContext,
  useMetricsViewMocks,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { invalidateMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { waitFor } from "@testing-library/svelte";
import { Settings } from "luxon";
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const NOW_TIME_RANGE = "7D as of now/D+1D";
const OTHER_TIME_RANGE = "7D as of latest/D";
// What getDefaultTimeRange picks for the 3 months of TIME_RANGE_SUMMARY.
const RILL_DEFAULT_TIME_RANGE = TimeRangePreset.QUARTER_TO_DATE;
const NO_TIME_METRICS_NAME = "AdBids_no_time";
// The field is left out rather than set to undefined, which the response decoder rejects.
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { timeDimension: _, ...NO_TIME_METRICS_SPEC } = AD_BIDS_METRICS_INIT;

const mocks = useMetricsViewMocks({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT_WITH_TIME,
  [NO_TIME_METRICS_NAME]: NO_TIME_METRICS_SPEC,
});

describe("TimeFilterManager", () => {
  let destroy: (() => void) | undefined;

  beforeEach(() => {
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, TIME_RANGE_SUMMARY);
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, RESOLVED_RILL_TIMES);
    queryClient.clear();
  });

  afterEach(() => {
    destroy?.();
    destroy = undefined;
    vi.restoreAllMocks();
  });

  function createManager(
    parent?: TimeFilterManager,
    yamlConfigProvider = new YAMLConfigProvider(),
    config: TimeFiltersConfig = {},
    metricsViewNames = [AD_BIDS_METRICS_NAME],
  ) {
    const rendered = renderInRuntimeContext(({ runtimeClient }) => {
      const metricsViewsProvider = new MetricsViewsProvider(
        runtimeClient,
        metricsViewNames,
      );
      const manager = new TimeFilterManager(
        runtimeClient,
        metricsViewsProvider,
        yamlConfigProvider,
        config,
        parent,
      );
      return { manager, metricsViewsProvider };
    });
    const destroyPrevious = destroy;
    destroy = () => {
      destroyPrevious?.();
      rendered.value.manager.cleanup();
      rendered.value.metricsViewsProvider.cleanup();
      rendered.destroy();
    };

    const resolvedRanges: string[] = [];
    rendered.value.manager.on("rangeResolved", () =>
      resolvedRanges.push(rendered.value.manager.timeRange!),
    );
    return { manager: rendered.value.manager, resolvedRanges };
  }

  function applyUrlTimeRange(manager: TimeFilterManager, timeRange: string) {
    manager.storeSync.setUrlParams(
      new URLSearchParams([[ExploreStateURLParams.TimeRange, timeRange]]),
    );
  }

  function expectInterval(
    manager: TimeFilterManager,
    { start, end }: { start?: string; end?: string },
  ) {
    expect(manager.interval?.start.toUTC().toISO()).toEqual(
      new Date(start!).toISOString(),
    );
    expect(manager.interval?.end.toUTC().toISO()).toEqual(
      new Date(end!).toISOString(),
    );
  }

  function countTimeRangesRequests(
    fetchSpy: ReturnType<typeof vi.spyOn>,
    metricsViewName?: string,
  ) {
    return fetchSpy.mock.calls.filter(([url, init]) => {
      if (!String(url).endsWith("/MetricsViewTimeRanges")) return false;
      if (!metricsViewName) return true;
      const body = (init as RequestInit | undefined)?.body;
      const bodyText =
        typeof body === "string"
          ? body
          : new TextDecoder().decode(body as ArrayBufferView);
      const request = JSON.parse(bodyText) as { metricsViewName?: string };
      return request.metricsViewName === metricsViewName;
    }).length;
  }

  it("re-resolves the interval when the metrics view is invalidated", async () => {
    const { manager, resolvedRanges } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );
    const resolvedCount = resolvedRanges.length;

    const reconciled = {
      start: "2024-03-26T00:00:00.000Z",
      end: "2024-04-02T00:00:00.000Z",
    };
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
      ...RESOLVED_RILL_TIMES,
      [DEFAULT_TIME_RANGE]: reconciled,
    });
    await invalidateMetricsViewData(queryClient, AD_BIDS_METRICS_NAME, false);

    await waitFor(() => expectInterval(manager, reconciled));
    expect(resolvedRanges.length).toBeGreaterThan(resolvedCount);
    expect(manager.timeRange).toEqual(DEFAULT_TIME_RANGE);
  });

  it("resolves the interval alongside a metrics view without a time dimension", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { manager } = createManager(undefined, undefined, undefined, [
      AD_BIDS_METRICS_NAME,
      NO_TIME_METRICS_NAME,
    ]);
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);

    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );
    // The runtime rejects resolving a range for a metrics view without a time dimension.
    expect(countTimeRangesRequests(fetchSpy, NO_TIME_METRICS_NAME)).toBe(0);
  });

  it("re-resolves the interval when the execution time changes", async () => {
    const { manager } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );

    // The mock ignores the execution time, so change what it serves to tell the requests apart.
    const asOfExecutionTime = {
      start: "2024-02-01T00:00:00.000Z",
      end: "2024-02-08T00:00:00.000Z",
    };
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
      ...RESOLVED_RILL_TIMES,
      [DEFAULT_TIME_RANGE]: asOfExecutionTime,
    });
    manager.setExecutionTime("2024-02-08T00:00:00.000Z");

    await waitFor(() => expectInterval(manager, asOfExecutionTime));
  });

  it("refetches when the same range anchored at now is applied again", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { manager } = createManager();
    applyUrlTimeRange(manager, NOW_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[NOW_TIME_RANGE]),
    );
    const requestCount = countTimeRangesRequests(fetchSpy);

    manager.onSelectRange(NOW_TIME_RANGE);

    await waitFor(() =>
      expect(countTimeRangesRequests(fetchSpy)).toBe(requestCount + 1),
    );
  });

  it("does not re-apply an unchanged range when other time params change in the url", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { manager, resolvedRanges } = createManager();
    applyUrlTimeRange(manager, NOW_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[NOW_TIME_RANGE]),
    );
    const requestCount = countTimeRangesRequests(fetchSpy);
    const resolvedCount = resolvedRanges.length;

    // E.g. navigating back after toggling the comparison.
    manager.storeSync.setUrlParams(
      new URLSearchParams([
        [ExploreStateURLParams.TimeRange, NOW_TIME_RANGE],
        [ExploreStateURLParams.ComparisonTimeRange, TimeComparisonOption.WEEK],
      ]),
    );
    expect(manager.showComparison).toBe(true);
    // A refetch would be issued, and resolve, asynchronously.
    await new Promise((resolve) => setTimeout(resolve, 50));

    expect(resolvedRanges.length).toBe(resolvedCount);
    expect(countTimeRangesRequests(fetchSpy)).toBe(requestCount);
  });

  it("does not refetch when the same range not anchored at now is applied again", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { manager, resolvedRanges } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );
    const requestCount = countTimeRangesRequests(fetchSpy);
    const resolvedCount = resolvedRanges.length;

    manager.onSelectRange(DEFAULT_TIME_RANGE);

    // Served from the cache, so it commits right away.
    expect(resolvedRanges.length).toBe(resolvedCount + 1);
    expect(countTimeRangesRequests(fetchSpy)).toBe(requestCount);
  });

  it("commits only the last of ranges applied in quick succession", async () => {
    const { manager, resolvedRanges } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    applyUrlTimeRange(manager, OTHER_TIME_RANGE);

    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[OTHER_TIME_RANGE]),
    );
    expect(resolvedRanges).not.toContain(DEFAULT_TIME_RANGE);
    expect(manager.timeRange).toEqual(OTHER_TIME_RANGE);
  });

  it("follows the selected grain without resolving the range again", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const { manager } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );
    const requestCount = countTimeRangesRequests(fetchSpy);
    const adjustedInterval = manager.adjustedInterval;

    manager.onSelectGrain(V1TimeGrain.TIME_GRAIN_HOUR);

    expect(manager.timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_HOUR);
    expect(manager.adjustedInterval).not.toBe(adjustedInterval);
    expect(countTimeRangesRequests(fetchSpy)).toBe(requestCount);
  });

  it("derives the comparison interval from the resolved interval and grain", async () => {
    const { manager } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );

    manager.onSelectComparisonRange(TimeComparisonOption.WEEK);
    // The week before the resolved interval.
    expect(manager.comparisonInterval?.start.toISO()).toEqual(
      "2024-03-18T00:00:00.000Z",
    );
    expect(manager.comparisonInterval?.end.toISO()).toEqual(
      "2024-03-25T00:00:00.000Z",
    );

    const adjustedComparisonInterval = manager.adjustedComparisonInterval;
    manager.onSelectGrain(V1TimeGrain.TIME_GRAIN_HOUR);
    expect(manager.adjustedComparisonInterval).not.toBe(
      adjustedComparisonInterval,
    );
  });

  it("inherits the parent's comparison range", async () => {
    const { manager: parent } = createManager();
    applyUrlTimeRange(parent, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(parent, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );
    parent.onSelectComparisonRange(TimeComparisonOption.WEEK);

    const { manager: child } = createManager(parent);
    child.onSelectComparisonRange(INHERIT_TIME_RANGE_ALIAS);
    expect(child.comparisonTimeRange).toEqual(TimeComparisonOption.WEEK);

    parent.onSelectComparisonRange(TimeComparisonOption.MONTH);
    expect(child.comparisonTimeRange).toEqual(TimeComparisonOption.MONTH);
  });

  it("pans back and forward by the width of the range, comparing against the previous period", async () => {
    const pannedLeft = "2024-03-18T00:00:00.000Z to 2024-03-25T00:00:00.000Z";
    const pannedRight = "2024-04-01T00:00:00.000Z to 2024-04-08T00:00:00.000Z";
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
      ...RESOLVED_RILL_TIMES,
      [pannedLeft]: {
        start: "2024-03-18T00:00:00.000Z",
        end: "2024-03-25T00:00:00.000Z",
      },
    });
    const { manager } = createManager();
    manager.storeSync.setUrlParams(
      new URLSearchParams([
        [ExploreStateURLParams.TimeRange, DEFAULT_TIME_RANGE],
        [ExploreStateURLParams.ComparisonTimeRange, TimeComparisonOption.MONTH],
      ]),
    );
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );

    manager.onPan("right");
    expect(manager.urlTimeRange).toEqual(pannedRight);

    manager.onPan("left");
    expect(manager.urlTimeRange).toEqual(pannedLeft);
    await waitFor(() => expect(manager.timeRange).toEqual(pannedLeft));

    // An absolute range compares against the previous period rather than the selected previous month.
    expect(manager.comparisonTimeRange).toEqual(
      TimeComparisonOption.CONTIGUOUS,
    );
    expect(manager.comparisonInterval?.start.toISO()).toEqual(
      "2024-03-11T00:00:00.000Z",
    );
    expect(manager.comparisonInterval?.end.toISO()).toEqual(
      "2024-03-18T00:00:00.000Z",
    );
  });

  it("restores the highlighted range as the instants it was written as", async () => {
    // Highlighted ranges used to be read in the browser's zone, ignoring the `Z`.
    const defaultZone = Settings.defaultZone;
    Settings.defaultZone = "Asia/Kathmandu";
    try {
      const { manager } = createManager();
      manager.storeSync.setUrlParams(
        new URLSearchParams([
          [ExploreStateURLParams.TimeRange, DEFAULT_TIME_RANGE],
          [
            ExploreStateURLParams.HighlightedTimeRange,
            "2024-03-26T00:00:00.000Z,2024-03-27T00:00:00.000Z",
          ],
        ]),
      );
      await waitFor(() =>
        expect(manager.lastDefinedScrubInterval?.start.toUTC().toISO()).toEqual(
          "2024-03-26T00:00:00.000Z",
        ),
      );
      expect(manager.lastDefinedScrubInterval?.end.toUTC().toISO()).toEqual(
        "2024-03-27T00:00:00.000Z",
      );
    } finally {
      Settings.defaultZone = defaultZone;
    }
  });

  describe("a range with a zone of its own", () => {
    const ZONED_TIME_RANGE = `${DEFAULT_TIME_RANGE} tz Asia/Kathmandu`;

    beforeEach(() => {
      mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
        ...RESOLVED_RILL_TIMES,
        // Midnight in Kathmandu, a week before latest/D+1D.
        [ZONED_TIME_RANGE]: {
          start: "2024-03-24T18:15:00.000Z",
          end: "2024-03-31T18:15:00.000Z",
          grain: V1TimeGrain.TIME_GRAIN_DAY,
        },
      });
    });

    it("applies its zone without a zone param, and pans in wall-clock times of that zone", async () => {
      const { manager } = createManager();
      applyUrlTimeRange(manager, ZONED_TIME_RANGE);
      await waitFor(() => expect(manager.timeRange).toEqual(ZONED_TIME_RANGE));
      expect(manager.timeZone).toEqual("Asia/Kathmandu");

      // The runtime reads absolute times in the selected zone, ignoring the `Z`.
      manager.onPan("right");
      expect(manager.urlTimeRange).toEqual(
        "2024-04-01T00:00:00.000Z to 2024-04-08T00:00:00.000Z",
      );
    });

    it("keeps a zone picked afterwards", async () => {
      const { manager } = createManager();
      const changes: string[] = [];
      manager.storeSync.on("internal-change", (params) =>
        changes.push(params.toString()),
      );
      applyUrlTimeRange(manager, ZONED_TIME_RANGE);
      await waitFor(() => expect(manager.timeRange).toEqual(ZONED_TIME_RANGE));

      manager.onSelectZone("America/New_York");
      await waitFor(() =>
        expect(changes.at(-1)).toContain(
          `${ExploreStateURLParams.TimeZone}=America%2FNew_York`,
        ),
      );
      expect(manager.timeZone).toEqual("America/New_York");
    });
  });

  it("resolves a custom comparison window and sends it as start and end", async () => {
    const { manager } = createManager();
    applyUrlTimeRange(manager, DEFAULT_TIME_RANGE);
    await waitFor(() =>
      expectInterval(manager, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
    );

    manager.onSelectComparisonRange(
      "2024-02-01T00:00:00.000Z to 2024-02-08T00:00:00.000Z",
    );
    expect(manager.isCustomComparison).toBe(true);
    const expected = {
      start: "2024-02-01T00:00:00.000Z",
      end: "2024-02-08T00:00:00.000Z",
    };
    expect(manager.comparisonTimeStart).toEqual(expected.start);
    expect(manager.comparisonTimeEnd).toEqual(expected.end);
    expect(
      mapTimeManagerComparisonRangeToV1TimeRange(
        manager,
        mapTimeManagerRangeToV1TimeRange(manager),
      ),
    ).toEqual(expected);

    // Older urls separate the custom window with a comma.
    manager.onSelectComparisonRange(`${expected.start},${expected.end}`);
    expect(manager.isCustomComparison).toBe(true);
    expect(manager.comparisonTimeStart).toEqual(expected.start);
  });

  it.each([
    {
      title: "an absolute range",
      timeRange: "2024-03-18 to 2024-03-25",
      config: {},
    },
    {
      title: "no range",
      timeRange: undefined,
      config: { skipDefaultTimeRange: true },
    },
  ])("reports a zone change on $title", async ({ timeRange, config }) => {
    if (timeRange) {
      mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
        ...RESOLVED_RILL_TIMES,
        [timeRange]: {
          start: "2024-03-18T00:00:00.000Z",
          end: "2024-03-25T00:00:00.000Z",
        },
      });
    }
    const { manager } = createManager(undefined, undefined, config);
    const changes: string[] = [];
    manager.storeSync.on("internal-change", (params) =>
      changes.push(params.toString()),
    );
    manager.storeSync.setUrlParams(
      new URLSearchParams(
        timeRange ? [[ExploreStateURLParams.TimeRange, timeRange]] : [],
      ),
    );
    await waitFor(() => expect(manager.ready).toBe(true));
    if (timeRange)
      await waitFor(() => expect(manager.timeRange).toEqual(timeRange));

    manager.onSelectZone("Asia/Kathmandu");
    await waitFor(() =>
      expect(changes.at(-1)).toContain(
        `${ExploreStateURLParams.TimeZone}=Asia%2FKathmandu`,
      ),
    );
    expect(manager.urlTimeRange).toEqual(timeRange);
  });

  describe("params missing from the url", () => {
    // Yaml defaults are applied before the url reaches the manager,
    // so a missing param means the rill default.
    function createManagerWithYamlDefaults(timeZones: string[]) {
      const yamlConfigProvider = new YAMLConfigProvider();
      yamlConfigProvider.update({
        defaultTimeRange: OTHER_TIME_RANGE,
        defaultTimeZone: "Asia/Kathmandu",
        timeZones,
      });
      return createManager(undefined, yamlConfigProvider).manager;
    }

    it("fall back to the rill defaults rather than the yaml defaults", async () => {
      const manager = createManagerWithYamlDefaults(YAML_TIME_ZONES);

      manager.storeSync.setUrlParams(
        new URLSearchParams([
          [
            ExploreStateURLParams.ComparisonTimeRange,
            TimeComparisonOption.WEEK,
          ],
        ]),
      );

      await waitFor(() =>
        expect(manager.urlTimeRange).toEqual(RILL_DEFAULT_TIME_RANGE),
      );
      // The first of the yaml time zones.
      expect(manager.timeZone).toEqual("UTC");
    });

    it("leaves out the zone only when it is the rill default", async () => {
      const manager = createManagerWithYamlDefaults([
        "America/New_York",
        "UTC",
      ]);
      const changes: string[] = [];
      manager.storeSync.on("internal-change", (params) =>
        changes.push(params.toString()),
      );

      manager.storeSync.setUrlParams(
        new URLSearchParams([
          [ExploreStateURLParams.TimeRange, DEFAULT_TIME_RANGE],
        ]),
      );
      await waitFor(() => expect(manager.timeZone).toEqual("America/New_York"));

      manager.onSelectZone("UTC");
      await waitFor(() =>
        expect(changes.at(-1)).toContain(
          `${ExploreStateURLParams.TimeZone}=UTC`,
        ),
      );

      manager.onSelectZone("America/New_York");
      await waitFor(() =>
        expect(changes.at(-1)).not.toContain(
          `${ExploreStateURLParams.TimeZone}=`,
        ),
      );
    });
  });

  describe("child following its parent", () => {
    async function createResolvedParent(params: string[][] = []) {
      const { manager: parent } = createManager();
      parent.storeSync.setUrlParams(
        new URLSearchParams([
          [ExploreStateURLParams.TimeRange, DEFAULT_TIME_RANGE],
          ...params,
        ]),
      );
      await waitFor(() =>
        expectInterval(parent, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
      );
      return parent;
    }

    function createChild(parent: TimeFilterManager, params: string) {
      const { manager: child } = createManager(parent);
      child.storeSync.setUrlParams(new URLSearchParams(params));
      return child;
    }

    it("follows the parent's range, also when it changes after the child resolved", async () => {
      const parent = await createResolvedParent();
      const child = createChild(parent, "tr=inherit&compare_tr=inherit");
      await waitFor(() =>
        expectInterval(child, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
      );
      expect(child.timeRange).toEqual(DEFAULT_TIME_RANGE);

      applyUrlTimeRange(parent, OTHER_TIME_RANGE);

      await waitFor(() =>
        expectInterval(child, RESOLVED_RILL_TIMES[OTHER_TIME_RANGE]),
      );
      expect(child.timeRange).toEqual(OTHER_TIME_RANGE);
      expect(child.urlTimeRange).toEqual(INHERIT_TIME_RANGE_ALIAS);
      // The picker locks the zone of a child that follows the parent's range.
      expect(child.inheritsTimeRange).toBe(true);
    });

    it("keeps its own range", async () => {
      const parent = await createResolvedParent();
      const child = createChild(parent, `tr=${OTHER_TIME_RANGE}`);
      await waitFor(() =>
        expectInterval(child, RESOLVED_RILL_TIMES[OTHER_TIME_RANGE]),
      );

      applyUrlTimeRange(parent, NOW_TIME_RANGE);
      await waitFor(() =>
        expectInterval(parent, RESOLVED_RILL_TIMES[NOW_TIME_RANGE]),
      );
      expect(child.timeRange).toEqual(OTHER_TIME_RANGE);
      // The picker lets a child with its own range pick a zone.
      expect(child.inheritsTimeRange).toBe(false);
    });

    it("follows the parent's comparison, whether it is shown or not", async () => {
      const parent = await createResolvedParent();
      const child = createChild(parent, "tr=inherit&compare_tr=inherit");
      await waitFor(() => expect(child.timeRange).toEqual(DEFAULT_TIME_RANGE));
      // An inherited comparison does not turn comparison on by itself.
      expect(child.showComparison).toBe(false);
      expect(get(child.getTimeControlStore()).comparisonInterval).toBe(
        undefined,
      );

      parent.onSelectComparisonRange(TimeComparisonOption.WEEK);
      expect(child.showComparison).toBe(true);
      expect(child.comparisonTimeRange).toEqual(TimeComparisonOption.WEEK);
      // Computed against the child's own interval.
      expect(
        get(child.getTimeControlStore()).comparisonInterval?.start.toISO(),
      ).toEqual("2024-03-18T00:00:00.000Z");

      parent.setShowComparison(false);
      expect(child.showComparison).toBe(false);
    });

    it("keeps its comparison off when the range is inherited without one", async () => {
      const parent = await createResolvedParent([
        [ExploreStateURLParams.ComparisonTimeRange, TimeComparisonOption.WEEK],
      ]);
      expect(parent.showComparison).toBe(true);

      const child = createChild(parent, "tr=inherit");
      await waitFor(() => expect(child.timeRange).toEqual(DEFAULT_TIME_RANGE));
      expect(child.showComparison).toBe(false);
    });

    it("follows the parent's zone unless it sets a zone of its own on a local range", async () => {
      const parent = await createResolvedParent([
        [ExploreStateURLParams.TimeZone, "Asia/Kathmandu"],
      ]);
      const inheriting = createChild(parent, "tr=inherit&compare_tr=inherit");
      const local = createChild(parent, `tr=${OTHER_TIME_RANGE}`);
      const ownZone = createChild(parent, `tr=${OTHER_TIME_RANGE}&tz=UTC`);
      await waitFor(() =>
        expect(inheriting.timeRange).toEqual(DEFAULT_TIME_RANGE),
      );
      await waitFor(() => expect(local.timeRange).toEqual(OTHER_TIME_RANGE));
      await waitFor(() => expect(ownZone.timeRange).toEqual(OTHER_TIME_RANGE));

      expect(inheriting.timeZone).toEqual("Asia/Kathmandu");
      expect(local.timeZone).toEqual("Asia/Kathmandu");
      expect(ownZone.timeZone).toEqual("UTC");
    });

    it("applies its own grain on an inherited range", async () => {
      const parent = await createResolvedParent();
      const withGrain = createChild(parent, "tr=inherit&grain=hour");
      const withoutGrain = createChild(parent, "tr=inherit");
      await waitFor(() =>
        expect(withGrain.timeRange).toEqual(DEFAULT_TIME_RANGE),
      );
      await waitFor(() =>
        expect(withoutGrain.timeRange).toEqual(DEFAULT_TIME_RANGE),
      );

      expect(parent.timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_DAY);
      expect(withGrain.timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(withoutGrain.timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("resolves against the parent's execution time", async () => {
      const parent = await createResolvedParent();
      const child = createChild(parent, "tr=inherit&compare_tr=inherit");
      await waitFor(() =>
        expectInterval(child, RESOLVED_RILL_TIMES[DEFAULT_TIME_RANGE]),
      );

      // The mock ignores the execution time, so change what it serves to tell the requests apart.
      const asOfExecutionTime = {
        start: "2024-02-01T00:00:00.000Z",
        end: "2024-02-08T00:00:00.000Z",
      };
      mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
        ...RESOLVED_RILL_TIMES,
        [DEFAULT_TIME_RANGE]: asOfExecutionTime,
      });
      parent.setExecutionTime("2024-02-08T00:00:00.000Z");

      await waitFor(() => expectInterval(child, asOfExecutionTime));
    });

    it("writes inherit back for the parts it follows", async () => {
      const parent = await createResolvedParent();
      const child = createChild(parent, "tr=inherit&compare_tr=inherit");
      await waitFor(() => expect(child.timeRange).toEqual(DEFAULT_TIME_RANGE));
      const changes: string[] = [];
      child.storeSync.on("internal-change", (params) =>
        changes.push(params.toString()),
      );

      child.onSelectRange(OTHER_TIME_RANGE);

      await waitFor(() =>
        expect(changes).toEqual([
          new URLSearchParams([
            [ExploreStateURLParams.TimeRange, OTHER_TIME_RANGE],
            [
              ExploreStateURLParams.ComparisonTimeRange,
              INHERIT_TIME_RANGE_ALIAS,
            ],
          ]).toString(),
        ]),
      );
    });
  });
});
