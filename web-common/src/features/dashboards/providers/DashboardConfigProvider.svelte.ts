import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  createQueryServiceResolveCanvas,
  createRuntimeServiceGetExplore,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

/**
 * Metrics view name and other yaml config provider based on dashboard type.
 */
export class DashboardConfigProvider {
  public readonly metricsViewsProvider: MetricsViewsProvider;
  public readonly yamlConfigProvider: YAMLConfigProvider;

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

      this.metricsViewsProvider.setMetricsViewNames(
        exploreSpec.metricsView ? [exploreSpec.metricsView] : [],
      );

      this.yamlConfigProvider.update({
        restrictedDimensions: exploreSpec.dimensions,
        primaryTimeDimension: metricsViewSpec.timeDimension,
        restrictedMeasures: exploreSpec.measures,

        defaultTimeRange: exploreSpec.defaultPreset?.timeRange,
        timeRanges: exploreSpec.timeRanges,
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
  public constructor(runtimeClient: RuntimeClient, canvasName: string) {
    super(runtimeClient);

    const resolveCanvasQuery = createQueryServiceResolveCanvas(runtimeClient, {
      canvas: canvasName,
    });
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

          defaultTimeRange: canvasSpec.defaultPreset?.timeRange,
          timeRanges: canvasSpec.timeRanges,
          timeZones: canvasSpec.timeZones,
        });
      },
    );

    this.cleanup = () => {
      resolveCanvasUnsub();
      this.metricsViewsProvider.cleanup();
      this.yamlConfigProvider.cleanup?.();
    };
  }
}
