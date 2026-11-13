import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import ReadonlyTimeFilters from "@rilldata/web-common/features/dashboards/time-controls/ReadonlyTimeFilters.svelte";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import {
  DEFAULT_TIME_RANGE,
  RESOLVED_RILL_TIMES,
  TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/time-controls/test/rill-time-mocks";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  renderInRuntimeContext,
  useMetricsViewMocks,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { render, waitFor } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

const mocks = useMetricsViewMocks({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT_WITH_TIME,
});

describe("ReadonlyTimeFilters", () => {
  mockAnimationsForComponentTesting();
  let destroy: (() => void) | undefined;

  beforeEach(() => {
    mocks.mockTimeRangeSummary(AD_BIDS_METRICS_NAME, TIME_RANGE_SUMMARY);
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, RESOLVED_RILL_TIMES);
    queryClient.clear();
  });

  afterEach(() => {
    destroy?.();
    destroy = undefined;
  });

  function createManager(params: string, parent?: TimeFilterManager) {
    const rendered = renderInRuntimeContext(({ runtimeClient }) => {
      const metricsViewsProvider = new MetricsViewsProvider(runtimeClient, [
        AD_BIDS_METRICS_NAME,
      ]);
      const manager = new TimeFilterManager(
        runtimeClient,
        metricsViewsProvider,
        new YAMLConfigProvider(),
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

    rendered.value.manager.storeSync.setUrlParams(new URLSearchParams(params));
    return rendered.value.manager;
  }

  async function createResolvedParent() {
    const parent = createManager(
      new URLSearchParams([
        ["tr", DEFAULT_TIME_RANGE],
        ["compare_tr", "rill-PW"],
      ]).toString(),
    );
    await waitFor(() => expect(parent.interval).toBeDefined());
    return parent;
  }

  function renderTimeFilters(timeFilterManager: TimeFilterManager) {
    return render(ReadonlyTimeFilters, { props: { timeFilterManager } });
  }

  it("shows the range and comparison of a top level manager, without its grain", async () => {
    const parent = await createResolvedParent();

    const { container } = renderTimeFilters(parent);

    await waitFor(() => {
      const chips = container.querySelectorAll(".chip");
      expect(chips).toHaveLength(2);
      expect(chips[0]).toHaveTextContent(/Last 7 days/);
      expect(chips[1]).toHaveTextContent(/vs/);
    });
  });

  it("shows nothing for a child that inherits its range and comparison", async () => {
    const parent = await createResolvedParent();
    const child = createManager("tr=inherit&compare_tr=inherit", parent);
    await waitFor(() => expect(child.interval).toBeDefined());

    const { container } = renderTimeFilters(child);

    expect(container.querySelectorAll(".chip")).toHaveLength(0);
  });

  it("shows only the grain a child sets on an inherited range", async () => {
    const parent = await createResolvedParent();
    const child = createManager(
      "tr=inherit&compare_tr=inherit&grain=hour",
      parent,
    );
    await waitFor(() => expect(child.interval).toBeDefined());

    const { container } = renderTimeFilters(child);

    const chips = container.querySelectorAll(".chip");
    expect(chips).toHaveLength(1);
    expect(chips[0]).toHaveTextContent(/by\s*hour/i);
  });

  it("shows the range a child sets itself", async () => {
    const parent = await createResolvedParent();
    const child = createManager(
      new URLSearchParams([["tr", "7D as of latest/D"]]).toString(),
      parent,
    );

    const { container } = renderTimeFilters(child);

    await waitFor(() => {
      const chips = container.querySelectorAll(".chip");
      expect(chips).toHaveLength(1);
      expect(chips[0]).toHaveTextContent(/Last 7 days/);
    });
  });
});
