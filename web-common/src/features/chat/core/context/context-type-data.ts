import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { prettyFormatV1TimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import type {
  InlineChatContext,
  InlineChatContextMetadata,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export enum ChatContextEntryType {
  Explore = "explore",
  MetricsView = "metrics_view",
  TimeRange = "time_range",
  Where = "where",
  Measures = "measures",
  Dimensions = "dimensions",
  DimensionValues = "dimension_values",
}

type ContextDataPerType = {
  getLabel: (ctx: InlineChatContext, meta: InlineChatContextMetadata) => string;
};

export const ContextTypeData: Partial<
  Record<ChatContextEntryType, ContextDataPerType>
> = {
  [ChatContextEntryType.MetricsView]: {
    getLabel: (ctx, meta) =>
      meta[ctx.values[0]]?.metricsViewSpec?.displayName || ctx.values[0],
  },

  [ChatContextEntryType.TimeRange]: {
    getLabel: (ctx) => {
      const [start, end] = ctx.values[0].split(" to ");
      return prettyFormatV1TimeRange({
        start: start,
        end: end ?? start,
      });
    },
  },

  [ChatContextEntryType.Measures]: {
    getLabel: (ctx, meta) => {
      const mes = meta[ctx.values[0]]?.measures[ctx.values[1]];
      return getMeasureDisplayName(mes) || ctx.values[1];
    },
  },

  [ChatContextEntryType.Dimensions]: {
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.values[0]]?.dimensions[ctx.values[1]];
      return getDimensionDisplayName(dim) || ctx.values[1];
    },
  },

  [ChatContextEntryType.DimensionValues]: {
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.values[0]]?.dimensions[ctx.values[1]];
      const dimLabel = getDimensionDisplayName(dim) || ctx.values[1];
      return dimLabel + ": " + ctx.values.slice(2).join(", ");
    },
  },
};
