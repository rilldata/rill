import { ChatContextEntryType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export const INLINE_CHAT_CONTEXT_TAG = "inline";
const INLINE_CHAT_CONTEXT_TYPE_ATTR = "data-context-type";
const INLINE_CHAT_CONTEXT_VALUE_ATTR = "data-context-value-";

export type InlineChatContext = {
  type: ChatContextEntryType;
  label: string;
  // Hierarchy of values.
  // EG, for ChatContextEntryType.DimensionValue, [metricsViewName, dimensionName, ...dimensionValues]
  values: string[];
};

export function inlineChatContextsAreEqual(
  ctx1: InlineChatContext,
  ctx2: InlineChatContext,
) {
  if (ctx1.type !== ctx2.type) return false;
  if (ctx1.values.length !== ctx2.values.length) return false;
  return ctx1.values.every((v, i) => v === ctx2.values[i]);
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

export function convertContextToAttrs(
  ctx: InlineChatContext,
): Record<string, string> {
  const valuesAttrEntries = ctx.values.map((v, i) => [
    INLINE_CHAT_CONTEXT_VALUE_ATTR + i,
    v,
  ]);
  return Object.fromEntries(
    [[INLINE_CHAT_CONTEXT_TYPE_ATTR, ctx.type]].concat(valuesAttrEntries),
  );
}

export function convertContextToInlinePrompt(ctx: InlineChatContext) {
  const parts = [`type="${ctx.type}"`];

  switch (ctx.type) {
    case ChatContextEntryType.MetricsView:
    case ChatContextEntryType.Explore:
      parts.push(`name="${ctx.values[0]}"`);
      break;

    case ChatContextEntryType.TimeRange:
      parts.push(`range="${ctx.values[0]}"`);
      break;

    case ChatContextEntryType.Dimensions:
    case ChatContextEntryType.Measures:
      parts.push(`metrics_view="${ctx.values[0]}"`);
      parts.push(`name="${ctx.values[1]}"`);
      break;
  }

  return `<${INLINE_CHAT_CONTEXT_TAG}>${parts.join(" ")}</${INLINE_CHAT_CONTEXT_TAG}>`;
}

const PARTS_REGEX = /\w+?="([^"]+?)"/g;

export function convertPromptValueToContext(
  contextValue: string,
): InlineChatContext | null {
  const parts = contextValue.matchAll(PARTS_REGEX);
  if (!parts) return null;
  const matchedParts = [...parts].map((p) => p[1]);
  const [type, ...values] = matchedParts;

  const entry = <InlineChatContext>{
    type,
    values,
    label: "", // TODO
  };

  return entry;
}
