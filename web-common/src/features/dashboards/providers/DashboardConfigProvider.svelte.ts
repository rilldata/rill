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

  // TODO: ensure cleanup is called
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
    this.cleanup = getExploreQuery.subscribe((getExploreResp) => {
      const exploreSpec =
        getExploreResp.data?.explore?.explore?.state?.validSpec ?? {};

      this.metricsViewsProvider.setMetricsViewNames(
        exploreSpec.metricsView ? [exploreSpec.metricsView] : [],
      );

      // this.yamlConfigProvider.update() // TODO: once we have this support for explore
    });
  }
}

export class CanvasDashboardConfigProvider extends DashboardConfigProvider {
  public constructor(runtimeClient: RuntimeClient, canvasName: string) {
    super(runtimeClient);

    const resolveCanvasQuery = createQueryServiceResolveCanvas(runtimeClient, {
      canvas: canvasName,
    });
    this.cleanup = resolveCanvasQuery.subscribe((resolveCanvasResp) => {
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
      this.yamlConfigProvider.update(
        defaultFilters,
        canvasSpec.pinnedFilters ?? [],
        canvasSpec.requiredFilters ?? [],
      );
    });
  }
}
