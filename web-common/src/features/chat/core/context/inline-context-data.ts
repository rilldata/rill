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
) {
  const optionsStore = getInlineChatContextOptions();

  return derived([optionsStore, searchTextStore], ([options, searchText]) => {
    const filterFunction = (label: string, value: string) =>
      searchText.length < 2 ||
      label.toLowerCase().includes(searchText.toLowerCase()) ||
      value.toLowerCase().includes(searchText.toLowerCase());

    return options
      .map((metricsViewOption) => {
        const filteredMeasures = metricsViewOption.measures.filter((m) =>
          filterFunction(m.label, m.values[1]),
        );

        const filteredDimensions = metricsViewOption.dimensions.filter((d) =>
          filterFunction(d.label, d.values[1]),
        );

        const metricsViewMatches = filterFunction(
          metricsViewOption.metricsViewContext.label,
          metricsViewOption.metricsViewContext.values[0],
        );

        if (
          !filteredMeasures.length &&
          !filteredDimensions.length &&
          !metricsViewMatches
        )
          return null;

        return {
          metricsViewContext: metricsViewOption.metricsViewContext,
          measures: filteredMeasures,
          dimensions: filteredDimensions,
        };
      })
      .filter(Boolean) as MetricsViewContextOption[];
  });
}
