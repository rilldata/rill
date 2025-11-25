import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { prettyFormatResolvedV1TimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export const INLINE_CHAT_CONTEXT_TAG = "inline";
export const INLINE_CHAT_TYPE_ATTR = "data-type";
export const INLINE_CHAT_METRICS_VIEW_ATTR = "data-metrics-view";
export const INLINE_CHAT_MEASURE_ATTR = "data-measure";
export const INLINE_CHAT_DIMENSION_ATTR = "data-dimension";
export const INLINE_CHAT_TIME_RANGE_ATTR = "data-time-range";

export enum ChatContextEntryType {
  Explore = "explore",
  MetricsView = "metricsView",
  TimeRange = "timeRange",
  Where = "where",
  Measure = "measure",
  Dimension = "dimension",
  DimensionValues = "dimensionValues",
}

export type InlineChatContext = {
  type: ChatContextEntryType;
  label?: string;
  metricsView?: string;
  measure?: string;
  dimension?: string;
  timeRange?: string;
  values?: string[];
};

export function inlineChatContextsAreEqual(
  ctx1: InlineChatContext,
  ctx2: InlineChatContext,
) {
  return (
    ctx1.type === ctx2.type &&
    ctx1.metricsView === ctx2.metricsView &&
    ctx1.measure === ctx2.measure &&
    ctx1.dimension === ctx2.dimension &&
    ctx1.timeRange === ctx2.timeRange
  );
}

export type InlineChatContextMetadata = Record<string, MetricsViewMetadata>;
export type MetricsViewMetadata = {
  metricsViewSpec: V1MetricsViewSpec;
  measures: Record<string, MetricsViewSpecMeasure>;
  dimensions: Record<string, MetricsViewSpecDimension>;
};

type ContextConfigPerType = {
  getLabel: (ctx: InlineChatContext, meta: InlineChatContextMetadata) => string;
};

export const InlineContextConfig: Partial<
  Record<ChatContextEntryType, ContextConfigPerType>
> = {
  [ChatContextEntryType.MetricsView]: {
    getLabel: (ctx, meta) =>
      meta[ctx.metricsView!]?.metricsViewSpec?.displayName || ctx.metricsView!,
  },

  [ChatContextEntryType.TimeRange]: {
    getLabel: (ctx) => {
      if (!ctx.timeRange) return "";
      const [start, end] = ctx.timeRange.split(" to ");
      return prettyFormatResolvedV1TimeRange({
        start: start,
        end: end ?? start,
      });
    },
  },

  [ChatContextEntryType.Measure]: {
    getLabel: (ctx, meta) => {
      const mes = meta[ctx.metricsView!]?.measures[ctx.measure!];
      return getMeasureDisplayName(mes) || ctx.measure!;
    },
  },

  [ChatContextEntryType.Dimension]: {
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.metricsView!]?.dimensions[ctx.dimension!];
      return getDimensionDisplayName(dim) || ctx.dimension!;
    },
  },

  [ChatContextEntryType.DimensionValues]: {
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.metricsView!]?.dimensions[ctx.dimension!];
      const dimLabel = getDimensionDisplayName(dim) || ctx.dimension!;
      return dimLabel + ": " + (ctx.values ?? []).join(", ");
    },
  },
};
