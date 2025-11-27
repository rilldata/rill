import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, type Readable } from "svelte/store";
import {
  InlineContextType,
  type InlineContext,
  type InlineContextMetadata,
  type MetricsViewMetadata,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { getLastUsedMetricsViewNameStore } from "@rilldata/web-common/features/chat/core/context/get-last-used-metrics-view.ts";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import { getActiveMetricsViewNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";

/**
 * Creates a store that contains a map of metrics view names to their metadata.
 * Each metrics view metadata has a reference to its spec, and a map of for measure and dimension spec by their names.
 */
export function getInlineChatContextMetadata() {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(),
    queryClient,
  );

  return derived(metricsViewsQuery, (metricsViewsResp) => {
    const metricsViews = metricsViewsResp.data ?? [];
    return Object.fromEntries(
      metricsViews.map((mv) => {
        const mvName = mv.meta?.name?.name ?? "";
        const metricsViewSpec = mv.metricsView?.state?.validSpec ?? {};

        const measures = Object.fromEntries(
          metricsViewSpec?.measures?.map((m) => [m.name!, m]) ?? [],
        );

        const dimensions = Object.fromEntries(
          metricsViewSpec?.dimensions?.map((d) => [d.name!, d]) ?? [],
        );

        return [
          mvName,
          <MetricsViewMetadata>{
            metricsViewSpec,
            measures,
            dimensions,
          },
        ];
      }),
    ) as InlineContextMetadata;
  });
}

export type MetricsViewContextOption = {
  metricsViewContext: InlineContext;
  recentlyUsed: boolean;
  currentlyActive: boolean;
  measures: InlineContext[];
  dimensions: InlineContext[];
};

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view.
 * 1st level: metrics view context options
 * 2nd level: measures and dimensions options for each metrics view
 */
export function getInlineChatContextOptions() {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(),
    queryClient,
  );

  return derived(metricsViewsQuery, (metricsViewsResp) => {
    const metricsViews = metricsViewsResp.data ?? [];
    return metricsViews.map((mv) => {
      const mvName = mv.meta?.name?.name ?? "";
      const metricsViewSpec = mv.metricsView?.state?.validSpec ?? {};
      const mvDisplayName = metricsViewSpec?.displayName || mvName;

      const measures =
        metricsViewSpec?.measures?.map(
          (m) =>
            <InlineContext>{
              type: InlineContextType.Measure,
              label: getMeasureDisplayName(m),
              metricsView: mvName,
              measure: m.name!,
            },
        ) ?? [];

      const dimensions =
        metricsViewSpec?.dimensions?.map(
          (d) =>
            <InlineContext>{
              type: InlineContextType.Dimension,
              label: getDimensionDisplayName(d),
              metricsView: mvName,
              dimension: d.name!,
            },
        ) ?? [];

      return {
        metricsViewContext: {
          type: InlineContextType.MetricsView,
          metricsView: mvName,
          label: mvDisplayName,
        },
        measures,
        dimensions,
      };
    });
  });
}

/**
 * Takes a store of search text and a conversation manager and returns a store of filtered metrics view context options.
 * The returned store contains options for the last used and active metrics views followed by options for all other valid metrics views.
 */
export function getInlineChatContextFilteredOptions(
  searchTextStore: Readable<string>,
  conversationManager: ConversationManager,
) {
  const optionsStore = getInlineChatContextOptions();
  const lastUsedMetricsViewStore =
    getLastUsedMetricsViewNameStore(conversationManager);
  const activeMetricsViewStore = getActiveMetricsViewNameStore();

  return derived(
    [
      optionsStore,
      lastUsedMetricsViewStore,
      activeMetricsViewStore,
      searchTextStore,
    ],
    ([options, lastUserMv, activeMv, searchText]) => {
      const filterFunction = (label: string, value: string) =>
        searchText.length === 0 ||
        label.toLowerCase().includes(searchText.toLowerCase()) ||
        value.toLowerCase().includes(searchText.toLowerCase());

      let lastUsedMvOption: MetricsViewContextOption | null = null;
      let activeMvOption: MetricsViewContextOption | null = null;

      const mvOptions = options.map((metricsViewOption) => {
        const filteredMeasures = metricsViewOption.measures.filter((m) =>
          filterFunction(m.label ?? "", m.measure ?? ""),
        );

        const filteredDimensions = metricsViewOption.dimensions.filter((d) =>
          filterFunction(d.label ?? "", d.dimension ?? ""),
        );

        const metricsViewName =
          metricsViewOption.metricsViewContext.metricsView;
        const metricsViewMatches = filterFunction(
          metricsViewOption.metricsViewContext.label,
          metricsViewName,
        );

        const shouldSkipMetricsView =
          !filteredMeasures.length &&
          !filteredDimensions.length &&
          !metricsViewMatches;
        if (shouldSkipMetricsView) return null;

        const recentlyUsed = lastUserMv === metricsViewName;
        const currentlyActive = activeMv === metricsViewName;
        const option = {
          metricsViewContext: metricsViewOption.metricsViewContext,
          recentlyUsed,
          currentlyActive,
          measures: filteredMeasures,
          dimensions: filteredDimensions,
        };

        if (recentlyUsed) lastUsedMvOption = option;
        if (currentlyActive) activeMvOption = option;

        if (recentlyUsed || currentlyActive) return null; // these are added explicitly
        return option;
      });

      // If the last used metrics view is the active metrics view, remove the active metrics view option from the list.
      if (lastUsedMvOption === activeMvOption) activeMvOption = null;

      return [lastUsedMvOption, activeMvOption, ...mvOptions].filter(
        Boolean,
      ) as MetricsViewContextOption[];
    },
  );
}
