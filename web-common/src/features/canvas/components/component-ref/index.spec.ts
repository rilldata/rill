import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type {
  V1CanvasItem,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { ComponentRefComponent } from "./index";

// The components build their own metrics views provider and time filter manager,
// which would subscribe to runtime queries; stub them with the parts the constructors touch.
vi.mock(
  "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts",
  () => ({
    MetricsViewsProvider: vi.fn(function () {
      return { setMetricsViewNames: vi.fn(), cleanup: vi.fn() };
    }),
  }),
);
vi.mock(
  "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts",
  async (importOriginal) => ({
    ...(await importOriginal<object>()),
    TimeFilterManager: vi.fn(function () {
      return {
        storeSync: { on: vi.fn(() => () => {}), setUrlParams: vi.fn() },
        getTimeControlStore: () => writable({}),
        cleanup: vi.fn(),
      };
    }),
  }),
);

describe("ComponentRefComponent", () => {
  it("syncs a bound metrics view into the shared metrics views provider", () => {
    const localExpressionFilters = {
      setParamForMetricsView: vi.fn(),
    };
    const parent = {
      allowUnvalidatedSpec: false,
      exportMode: writable(false),
      client: {},
      dashboardProvider: { yamlConfigProvider: {} },
      timeFilterManager: {},
      expressionFilterManager: {
        createLocalFilterStore: vi.fn(() => localExpressionFilters),
      },
    } as unknown as CanvasEntity;
    const resource = {
      meta: { name: { name: "area_chart" } },
      component: {
        state: {
          validSpec: {
            renderer: "custom_chart",
            params: [{ name: "metrics_view", type: "metrics_view" }],
          },
        },
      },
    } as V1Resource;
    const item = {
      component: "area_chart",
      params: { metrics_view: "bids_metrics" },
    } as V1CanvasItem;

    const component = new ComponentRefComponent(resource, parent, [], item);

    expect(component.metricsViewName).toBe("bids_metrics");
    // The local filter and time managers share this provider, so both follow.
    expect(component.metricsViewsProvider).toBe(
      vi.mocked(MetricsViewsProvider).mock.results[0].value,
    );
    expect(
      component.metricsViewsProvider.setMetricsViewNames,
    ).toHaveBeenCalledWith(["bids_metrics"]);
    expect(get(component.specStore).metrics_view).toBe("bids_metrics");

    component.destroy();
  });
});
