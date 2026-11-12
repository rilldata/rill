import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  createQueryServiceResolveCanvas,
  createRuntimeServiceGetExplore,
} from "@rilldata/web-common/runtime-client";
import { getResolveCanvasRequest } from "@rilldata/web-common/features/canvas/selector";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { normalizeWeekday } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import { Settings } from "luxon";

/**
 * Metrics view name and other yaml config provider based on dashboard type.
 */
export class DashboardConfigProvider {
  public readonly metricsViewsProvider: MetricsViewsProvider;
  public readonly yamlConfigProvider: YAMLConfigProvider;
  public defaultUrlParams: URLSearchParams = $state(new URLSearchParams());

  public cleanup: (() => void) | undefined = undefined;

  public constructor(runtimeClient: RuntimeClient) {
    this.metricsViewsProvider = new MetricsViewsProvider(runtimeClient, []);
    this.yamlConfigProvider = new YAMLConfigProvider();
  }
}

export class ExploreDashboardConfigProvider extends DashboardConfigProvider {
  public constructor(runtimeClient: RuntimeClient, exploreName: string) {
    super(runtimeClient);

    const getExploreQuery = createRuntimeServiceGetExplore(runtimeClient, {
      name: exploreName,
    });
    const getExploreUnsub = getExploreQuery.subscribe((getExploreResp) => {
      const metricsViewSpec =
        getExploreResp.data?.metricsView?.metricsView?.state?.validSpec ?? {};
      const exploreSpec =
        getExploreResp.data?.explore?.explore?.state?.validSpec ?? {};

      // Prefer the resource's own name over the explore's reference to it. Resource names are
      // case-insensitive in the runtime, so the two can differ in casing, and the rest of the
      // explore identifies the metrics view by the resource name.
      const metricsViewName =
        getExploreResp.data?.metricsView?.meta?.name?.name ??
        exploreSpec.metricsView;
      this.metricsViewsProvider.setMetricsViewNames(
        metricsViewName ? [metricsViewName] : [],
      );

      this.yamlConfigProvider.update({
        restrictedDimensions: exploreSpec.dimensions,
        primaryTimeDimension: metricsViewSpec.timeDimension,
        restrictedMeasures: exploreSpec.measures,

        defaultTimeRange: exploreSpec.defaultPreset?.timeRange,
        timeRanges: exploreSpec.timeRanges,
        allowCustomTimeRange: exploreSpec.allowCustomTimeRange,
        defaultTimeZone: exploreSpec.defaultPreset?.timezone,
        timeZones: exploreSpec.timeZones,
      });
    });

    this.cleanup = () => {
      getExploreUnsub();
      this.metricsViewsProvider.cleanup();
      this.yamlConfigProvider.cleanup?.();
    };
  }
}

export class CanvasDashboardConfigProvider extends DashboardConfigProvider {
  public constructor(
    runtimeClient: RuntimeClient,
    canvasName: string,
    allowUnvalidatedSpec = false,
  ) {
    super(runtimeClient);

    // The editor saves the default, pinned and required filters back to the YAML,
    // so it must read them from the full canvas, including content hidden from the current user.
    const resolveCanvasQuery = createQueryServiceResolveCanvas(
      runtimeClient,
      getResolveCanvasRequest(canvasName, allowUnvalidatedSpec),
    );
    // TODO: figure out a better way of handling this property
    //       when it's not consistent across all metrics views.
    const specsLoadedUnsub = this.metricsViewsProvider.on(
      "specs-loaded",
      () => {
        const firstSpecWithTimeSeries =
          this.metricsViewsProvider.metricsViewNames
            .map((name) => this.metricsViewsProvider.specs[name])
            .find((spec) => spec?.timeDimension);
        if (!firstSpecWithTimeSeries) return;

        Settings.defaultWeekSettings = {
          firstDay: normalizeWeekday(firstSpecWithTimeSeries.firstDayOfWeek),
          weekend: [6, 7],
          minimalDays: 4,
        };
      },
    );

    const resolveCanvasUnsub = resolveCanvasQuery.subscribe(
      (resolveCanvasResp) => {
        const canvasSpec =
          resolveCanvasResp.data?.canvas?.canvas?.state?.validSpec ?? {};

        this.metricsViewsProvider.setMetricsViewNames(
          Object.keys(resolveCanvasResp.data?.referencedMetricsViews ?? {}),
        );

        const defaultFilters = Object.fromEntries(
          Object.entries(canvasSpec.defaultPreset?.filterExpr ?? {}).map(
            ([mv, sqlFilter]) => [mv, sqlFilter.expression],
          ),
        );
        this.yamlConfigProvider.update({
          defaultFilters,
          pinnedFilters: canvasSpec.pinnedFilters,
          requiredFilters: canvasSpec.requiredFilters,

          // Without `defaults.time_range`, a canvas opens on the first `time_ranges` entry.
          defaultTimeRange:
            canvasSpec.defaultPreset?.timeRange ??
            canvasSpec.timeRanges?.[0]?.range,
          timeRanges: canvasSpec.timeRanges,
          allowCustomTimeRange: canvasSpec.allowCustomTimeRange,
          timeZones: canvasSpec.timeZones,
        });
      },
    );

    this.cleanup = () => {
      specsLoadedUnsub();
      resolveCanvasUnsub();
      this.metricsViewsProvider.cleanup();
      this.yamlConfigProvider.cleanup?.();
    };
  }
}
