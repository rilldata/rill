import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { TimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeManager.svelte.ts";
import { ComparisonTimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/ComparisonTimeRangeManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils.ts";

export class TimeFilterManager {
  public timeRangeManager: TimeRangeManager;
  public comparisonTimeRangeManager: ComparisonTimeRangeManager;

  public curStateParams = $state(new URLSearchParams());
  public curSetParams = $state(new URLSearchParams());

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

  public createListener() {
    this.timeRangeManager.createListener();
    this.comparisonTimeRangeManager.createListener();

    $effect(() => {
      const newParams = new URLSearchParams();
      this.timeRangeManager.applyFilterToParams(newParams);
      this.comparisonTimeRangeManager.applyFilterToParams(newParams);
      if (newParams.toString() === this.curStateParams.toString()) return;
      this.curStateParams = newParams;
    });
  }

  public setUrlParams(urlParams: URLSearchParams) {
    this.timeRangeManager.setUrlParams(urlParams);
    this.comparisonTimeRangeManager.setUrlParams(urlParams);

    const newSetParams = new URLSearchParams(
      this.timeRangeManager.curSetParams,
    );
    copyParamsToTarget(
      this.comparisonTimeRangeManager.curSetParams,
      newSetParams,
    );
    this.curSetParams = newSetParams;
  }
}
