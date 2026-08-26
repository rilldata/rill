import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { TimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeManager.svelte.ts";
import { ComparisonTimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/ComparisonTimeRangeManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

export class TimeFilterManager {
  public timeRangeManager: TimeRangeManager;
  public comparisonTimeRangeManager: ComparisonTimeRangeManager;

  public constructor(
    runtimeClient: RuntimeClient,
    metricsViewsProvider: MetricsViewsProvider,
    yamlConfigProvider: YAMLConfigProvider,
    allowCustomTimeRange: boolean,
  ) {
    this.timeRangeManager = new TimeRangeManager(
      runtimeClient,
      metricsViewsProvider,
    );
    this.comparisonTimeRangeManager = new ComparisonTimeRangeManager(
      yamlConfigProvider,
      this.timeRangeManager,
      allowCustomTimeRange,
    );
  }
}
