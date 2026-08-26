import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { TimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeManager.svelte.ts";
import { ComparisonTimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/ComparisonTimeRangeManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { page } from "$app/state";
import { goto } from "$app/navigation";

export class TimeFilterManager {
  public timeRangeManager: TimeRangeManager;
  public comparisonTimeRangeManager: ComparisonTimeRangeManager;

  public constructor(
    runtimeClient: RuntimeClient,
    private readonly metricsViewsProvider: MetricsViewsProvider,
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

  public syncWithUrl(
    defaultUrlParamsGetter: () => URLSearchParams | undefined,
  ) {
    let lock = false;

    $effect(() => {
      if (!this.metricsViewsProvider.ready || lock) return;
      lock = true;

      const newUrlSearch = new URLSearchParams(page.url.searchParams);
      if (defaultUrlParamsGetter) {
        const defaultUrlParams = defaultUrlParamsGetter();
        if (defaultUrlParams) {
          defaultUrlParams.forEach((value, key) => {
            newUrlSearch.set(key, value);
          });
        }
      }

      this.timeRangeManager.setUrlParams(newUrlSearch);
      this.comparisonTimeRangeManager.setUrlParams(newUrlSearch);

      lock = false;
    });

    $effect(() => {
      if (!this.metricsViewsProvider.ready || lock) return;
      lock = true;

      const newUrl = new URL(page.url);
      this.timeRangeManager.applyFilterToParams(newUrl.searchParams);
      this.comparisonTimeRangeManager.applyFilterToParams(newUrl.searchParams);

      if (newUrl.search === page.url.search) {
        lock = false;
        return;
      }

      void goto(newUrl).then(
        () => (lock = false),
        () => (lock = false),
      );
    });
  }
}
