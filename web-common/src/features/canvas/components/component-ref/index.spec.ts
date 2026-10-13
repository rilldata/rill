import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type {
  V1CanvasItem,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import { ComponentRefComponent } from "./index";

describe("ComponentRefComponent", () => {
  it("syncs a bound metrics view into the existing local filter and time state", () => {
    const setMetricsViewNames = vi.fn();
    const localTimeControls = {
      metricsViewName: undefined as string | undefined,
      onUrlChange: vi.fn(),
    };
    const localExpressionFilters = {
      setParamForMetricsView: vi.fn(),
      metricsViewsProvider: {
        setMetricsViewNames,
        cleanup: vi.fn(),
      },
      yamlConfigProvider: { cleanup: vi.fn() },
    };
    const parent = {
      allowUnvalidatedSpec: false,
      exportMode: writable(false),
      timeManager: {
        createLocalTimeState: vi.fn(() => localTimeControls),
      },
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
    expect(setMetricsViewNames).toHaveBeenCalledWith(["bids_metrics"]);
    expect(localTimeControls.metricsViewName).toBe("bids_metrics");
    expect(get(component.specStore).metrics_view).toBe("bids_metrics");

    component.destroy();
  });
});
