import { createQuery } from "@tanstack/svelte-query";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, type Readable } from "svelte/store";
import { ChatContextEntryType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type {
  InlineChatContext,
  MetricsViewMetadata,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { getLastUsedMetricsViewStore } from "@rilldata/web-common/features/chat/core/get-last-used-metrics-view.ts";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import { getActiveMetricsViewNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";

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
    );
  });
}

type MetricsViewContextOption = {
  metricsViewContext: InlineChatContext;
  recentlyUsed: boolean;
  currentlyActive: boolean;
  measures: InlineChatContext[];
  dimensions: InlineChatContext[];
};

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
        metricsViewSpec?.measures?.map((m) => ({
          type: ChatContextEntryType.Measures,
          label: getMeasureDisplayName(m),
          values: [mvName, m.name!],
        })) ?? [];

      const dimensions =
        metricsViewSpec?.dimensions?.map((d) => ({
          type: ChatContextEntryType.Dimensions,
          label: getDimensionDisplayName(d),
          values: [mvName, d.name!],
        })) ?? [];

      return {
        metricsViewContext: {
          type: ChatContextEntryType.MetricsView,
          label: mvDisplayName,
          values: [mvName],
        },
        measures,
        dimensions,
      };
    });
  });
}

export function getInlineChatContextFilteredOptions(
  searchTextStore: Readable<string>,
  conversationManager: ConversationManager,
) {
  const optionsStore = getInlineChatContextOptions();
  const lastUsedMetricsViewStore =
    getLastUsedMetricsViewStore(conversationManager);
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
        searchText.length < 2 ||
        label.toLowerCase().includes(searchText.toLowerCase()) ||
        value.toLowerCase().includes(searchText.toLowerCase());

      let lastUsedMvOption: MetricsViewContextOption | null = null;
      let activeMvOption: MetricsViewContextOption | null = null;

      const mvOptions = options.map((metricsViewOption) => {
        const filteredMeasures = metricsViewOption.measures.filter((m) =>
          filterFunction(m.label, m.values[1]),
        );

        const filteredDimensions = metricsViewOption.dimensions.filter((d) =>
          filterFunction(d.label, d.values[1]),
        );

        const metricsViewName = metricsViewOption.metricsViewContext.values[0];
        const metricsViewMatches = filterFunction(
          metricsViewOption.metricsViewContext.label,
          metricsViewName,
        );

        if (
          !filteredMeasures.length &&
          !filteredDimensions.length &&
          !metricsViewMatches
        )
          return null;

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

      if (lastUsedMvOption === activeMvOption) activeMvOption = null;

      return [lastUsedMvOption, activeMvOption, ...mvOptions].filter(
        Boolean,
      ) as MetricsViewContextOption[];
    },
  );
}
