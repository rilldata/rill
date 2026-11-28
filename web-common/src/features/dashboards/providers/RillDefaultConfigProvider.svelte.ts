import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  getDefaultTimeRange,
  getDefaultTimeZone,
} from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state.ts";

/**
 * Rill's own defaults, used for params the url leaves out.
 * Unlike YAMLConfigProvider, which only holds what the yaml sets,
 * these are derived from the data and the spec.
 *
 * Yaml defaults are applied before the url reaches the filter managers:
 * explore merges them in DashboardStateDataLoader and cleans the url against these rill defaults,
 * and canvas redirects an empty url to its yaml defaults.
 * So a param missing from the url means the rill default.
 */
export class RillDefaultConfigProvider {
  /**
   * Based on the time range summary and the smallest grain of the metrics views.
   * Undefined until the time range summary loads, or without a time dimension.
   */
  public readonly defaultTimeRange: string | undefined;
  /**
   * The first of the dashboard's time zones, UTC when none are set.
   */
  public readonly defaultTimeZone: string;

  public constructor(
    metricsViewsProvider: MetricsViewsProvider,
    yamlConfigProvider: YAMLConfigProvider,
  ) {
    this.defaultTimeRange = $derived(
      getDefaultTimeRange(
        metricsViewsProvider.largestMinTimeGrain,
        metricsViewsProvider.timeRangeSummary,
      ),
    );
    this.defaultTimeZone = $derived(
      getDefaultTimeZone(yamlConfigProvider.timeZones),
    );
  }
}
