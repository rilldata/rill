import { page } from "$app/state";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { goto } from "$app/navigation";

export class ExpressionFilterURLSync {
  public constructor(expressionFilterManager: ExpressionFilterManager) {
    $effect(() => expressionFilterManager.setUrlParams(page.url.searchParams));

    $effect(() => {
      const newUrl = new URL(page.url);

      expressionFilterManager.metricsViewsProvider.metricsViewNames.forEach(
        (mvName) => {
          const key = `${ExploreStateURLParams.Filters}.${mvName}`;
          if (mvName in expressionFilterManager.exprByMetricsView) {
            const param = convertExpressionToFilterParam(
              expressionFilterManager.exprByMetricsView[mvName],
              expressionFilterManager.inList,
            );
            newUrl.searchParams.set(key, param);
          } else {
            newUrl.searchParams.delete(key);
          }
        },
      );

      if (newUrl.search === page.url.search) return;
      void goto(newUrl);
    });
  }
}
