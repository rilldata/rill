import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { TimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeManager.svelte.ts";
import { ComparisonTimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/ComparisonTimeRangeManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils.ts";
import type { UrlParamsStore } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";

export class TimeFilterManager implements UrlParamsStore {
  public timeRangeManager: TimeRangeManager;
  public comparisonTimeRangeManager: ComparisonTimeRangeManager;

  public curParams = $state(new URLSearchParams());

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

  public setUrlParams(urlParams: URLSearchParams) {
    this.timeRangeManager.setUrlParams(urlParams);
    this.comparisonTimeRangeManager.setUrlParams(urlParams);

    const newSetParams = new URLSearchParams(this.timeRangeManager.curParams);
    copyParamsToTarget(this.comparisonTimeRangeManager.curParams, newSetParams);
    this.curParams = newSetParams;
  }

  public applyFilterToParams(urlParams: URLSearchParams) {
    this.timeRangeManager.applyFilterToParams(urlParams);
    this.comparisonTimeRangeManager.applyFilterToParams(urlParams);
  }
}
