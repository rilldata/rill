import {
  type ChatContextEntry,
  ChatContextEntryType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  getQueryServiceMetricsViewAggregationQueryOptions,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getContextDimensionValuesQueryOptions(
  ctxStore: Readable<ChatContextEntry | null>,
  searchTextStore: Readable<string>,
) {
  const exploreNameStore = getExploreNameStore();

  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
    queryClient,
  );

  return derived(
    [runtime, ctxStore, validSpecQuery, searchTextStore],
    ([{ instanceId }, ctx, validSpecResp, searchText]) => {
      const dimensionName =
        ctx?.type === ChatContextEntryType.DimensionValue ? ctx?.value : "";

      const metricsViewName =
        validSpecResp.data?.exploreSpec?.metricsView ?? "";
      let where: V1Expression | undefined = undefined;
      if (searchText.length) {
        const addNull = "null".includes(searchText);
        where = addNull
          ? createInExpression(dimensionName, [null])
          : createLikeExpression(dimensionName, `%${searchText}%`);
      }

      return getQueryServiceMetricsViewAggregationQueryOptions(
        instanceId,
        metricsViewName,
        {
          dimensions: [{ name: dimensionName }],
          limit: "50",
          offset: "0",
          sort: [{ name: dimensionName }],
          where,
        },
        {
          query: {
            enabled: !!dimensionName,
            select: (data) => {
              return (
                data.data?.map((d) => ({
                  label: d[dimensionName] as string,
                  value: d[dimensionName] as string, // TODO: non-string values
                })) ?? []
              );
            },
          },
        },
      );
    },
  );
}
