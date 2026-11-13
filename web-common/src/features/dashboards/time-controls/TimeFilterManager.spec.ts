import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { INHERIT_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
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
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const NOW_TIME_RANGE = "7D as of now/D+1D";
const OTHER_TIME_RANGE = "7D as of latest/D";
// What getDefaultTimeRange picks for the 3 months of TIME_RANGE_SUMMARY.
const RILL_DEFAULT_TIME_RANGE = TimeRangePreset.QUARTER_TO_DATE;

const mocks = useMetricsViewMocks({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT_WITH_TIME,
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
  ) {
    const rendered = renderInRuntimeContext(({ runtimeClient }) => {
      const metricsViewsProvider = new MetricsViewsProvider(runtimeClient, [
        AD_BIDS_METRICS_NAME,
      ]);
      const manager = new TimeFilterManager(
        runtimeClient,
        metricsViewsProvider,
        yamlConfigProvider,
        {},
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

  function countTimeRangesRequests(fetchSpy: ReturnType<typeof vi.spyOn>) {
    return fetchSpy.mock.calls.filter(([url]) =>
      String(url).endsWith("/MetricsViewTimeRanges"),
    ).length;
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

    it("follows the parent's zone only while the range is inherited", async () => {
      const parent = await createResolvedParent([
        [ExploreStateURLParams.TimeZone, "Asia/Kathmandu"],
      ]);
      const inheriting = createChild(parent, "tr=inherit&compare_tr=inherit");
      const local = createChild(parent, `tr=${OTHER_TIME_RANGE}`);
      await waitFor(() =>
        expect(inheriting.timeRange).toEqual(DEFAULT_TIME_RANGE),
      );
      await waitFor(() => expect(local.timeRange).toEqual(OTHER_TIME_RANGE));

      expect(inheriting.timeZone).toEqual("Asia/Kathmandu");
      expect(local.timeZone).toEqual("UTC");
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
