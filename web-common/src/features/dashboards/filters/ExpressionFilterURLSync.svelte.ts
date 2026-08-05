import { page } from "$app/state";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { goto } from "$app/navigation";

export class ExpressionFilterURLSync {
  private prevUrlSearch = "";

  public constructor(expressionFilterManager: ExpressionFilterManager) {
    $effect(() => expressionFilterManager.setUrlParams(page.url.searchParams));
    $effect(() => {
      const newUrlSearchParams = new URLSearchParams(page.url.searchParams);

      expressionFilterManager.metricsViewsProvider.metricsViewNames.forEach(
        (mvName) => {
          const key = `${ExploreStateURLParams.Filters}.${mvName}`;
          if (mvName in expressionFilterManager.exprByMetricsView) {
            const param = convertExpressionToFilterParam(
              expressionFilterManager.exprByMetricsView[mvName],
              expressionFilterManager.inList,
            );
            newUrlSearchParams.set(key, param);
          } else {
            newUrlSearchParams.delete(key);
          }
        },
      );

      const newUrlSearch = newUrlSearchParams.toString();
      if (newUrlSearch === this.prevUrlSearch) return;

      this.prevUrlSearch = newUrlSearch;
      void goto(`?${newUrlSearch}`);
    });
  }
}
