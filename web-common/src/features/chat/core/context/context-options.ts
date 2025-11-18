import {
  type ChatContextEntry,
  ChatContextEntryType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { getContextDimensionValuesQueryOptions } from "@rilldata/web-common/features/chat/core/context/get-context-dimension-values.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { getValidDashboardsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getContextOptions(
  ctxStore: Readable<ChatContextEntry>,
  searchTextStore: Readable<string>,
) {
  const exploreNameStore = getExploreNameStore();

  const exploresSpecQuery = createQuery(
    getValidDashboardsQueryOptions(),
    queryClient,
  );
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
    queryClient,
  );

  const dimensionValuesQuery = createQuery(
    getContextDimensionValuesQueryOptions(ctxStore, searchTextStore),
    queryClient,
  );

  const filterFunction = (
    { label, value }: { label: string; value: string },
    st: string,
  ) =>
    st.length < 2 ||
    label.toLowerCase().includes(st.toLowerCase()) ||
    value.toLowerCase().includes(st.toLowerCase());

  return derived(
    [exploresSpecQuery, validSpecQuery, dimensionValuesQuery, searchTextStore],
    ([exploresSpecResp, validSpecResp, dimensionValuesResp, searchText]) => {
      const exploreOptions =
        exploresSpecResp.data
          ?.map((r) => {
            const exploreName = r.meta?.name?.name ?? "";
            const exploreSpec = r.explore?.state?.validSpec ?? {};
            return {
              value: exploreName,
              label: exploreSpec.displayName || exploreName,
            };
          })
          .filter((o) => filterFunction(o, searchText)) ?? [];

      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};

      const measuresOptions =
        metricsViewSpec.measures
          ?.map((m) => ({
            value: m.name ?? "",
            label: getMeasureDisplayName(m),
          }))
          .filter(
            (o) =>
              exploreSpec.measures?.includes(o.value) &&
              filterFunction(o, searchText),
          ) ?? [];
      const dimensionsOptions =
        metricsViewSpec.dimensions
          ?.map((d) => ({
            value: d.name ?? "",
            label: getDimensionDisplayName(d),
          }))
          .filter(
            (o) =>
              exploreSpec.dimensions?.includes(o.value) &&
              filterFunction(o, searchText),
          ) ?? [];

      const dimensionValuesOptions = dimensionValuesResp.data ?? [];

      return {
        [ChatContextEntryType.Measures]: measuresOptions,
        [ChatContextEntryType.Dimensions]: dimensionsOptions,
        [ChatContextEntryType.DimensionValue]: dimensionValuesOptions,
        [ChatContextEntryType.Explore]: exploreOptions,
      };
    },
  );
}
