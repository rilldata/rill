import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { prettyFormatV1TimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  type RuntimeServiceCompleteBody,
  type V1Expression,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
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

export type ContextRecord = {
  explore?: string;
  metricsView?: string;
  dimensions?: string[];
  measures?: string[];
  where?: V1Expression;
  timeRange?: V1TimeRange;
  dimensionValue?: string;
};

type ContextDataPerType<K extends keyof ContextRecord = keyof ContextRecord> = {
  key: K;
  icon: typeof ExploreIcon;
  serializer: (value: ContextRecord[K]) => RuntimeServiceCompleteBody;
  deserializer: (rawContext: Record<string, unknown>) => ContextRecord[K];
  getLabel: (ctx: InlineChatContext, meta: InlineChatContextMetadata) => string;
};

export const ContextTypeData: Partial<
  Record<ChatContextEntryType, ContextDataPerType>
> = {
  [ChatContextEntryType.Explore]: <ContextDataPerType<"explore">>{
    key: "explore",
    icon: ExploreIcon,
    serializer: (explore) => ({ explore }),
    deserializer: (rawContext) => rawContext.explore,
    getLabel: (ctx) => ctx.values[0],
  },

  [ChatContextEntryType.MetricsView]: <ContextDataPerType<"metricsView">>{
    key: "metricsView",
    icon: ExploreIcon,
    serializer: (metricsView) => ({ metricsView }),
    deserializer: (rawContext) => rawContext.metrics_view,
    getLabel: (ctx, meta) =>
      meta[ctx.values[0]]?.metricsViewSpec?.displayName || ctx.values[0],
  },

  [ChatContextEntryType.TimeRange]: <ContextDataPerType<"timeRange">>{
    key: "timeRange",
    icon: Calendar,
    serializer: (timeRange) => ({
      timeStart: timeRange?.start,
      timeEnd: timeRange?.end,
    }),
    deserializer: (rawContext) =>
      rawContext.time_start && rawContext.time_end
        ? {
            start: rawContext.time_start,
            end: rawContext.time_end,
          }
        : undefined,
    getLabel: (ctx) => {
      const [start, end] = ctx.values[0].split(" to ");
      return prettyFormatV1TimeRange({
        start: start,
        end: end ?? start,
      });
    },
  },

  [ChatContextEntryType.Where]: <ContextDataPerType<"where">>{
    key: "where",
    icon: Filter,
    serializer: (where) => ({ where }),
    deserializer: (rawContext) =>
      mapResolverExpressionToV1Expression(rawContext.where as any),
    getLabel: (ctx) => ctx.values[0],
  },

  [ChatContextEntryType.Measures]: <ContextDataPerType<"measures">>{
    key: "measures",
    icon: Compare,
    serializer: (measures) => ({ measures }),
    deserializer: (rawContext) =>
      (rawContext.measures as Array<unknown>)?.length
        ? rawContext.measures
        : undefined,
    getLabel: (ctx, meta) => {
      const mes = meta[ctx.values[0]]?.measures[ctx.values[1]];
      return getMeasureDisplayName(mes) || ctx.values[1];
    },
  },

  [ChatContextEntryType.Dimensions]: <ContextDataPerType<"dimensions">>{
    key: "dimensions",
    icon: Compare,
    serializer: (dimensions) => ({ dimensions }),
    deserializer: (rawContext) =>
      (rawContext.dimensions as Array<unknown>)?.length
        ? rawContext.dimensions
        : undefined,
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.values[0]]?.dimensions[ctx.values[1]];
      return getDimensionDisplayName(dim) || ctx.values[1];
    },
  },

  [ChatContextEntryType.DimensionValues]: <
    ContextDataPerType<"dimensionValue">
  >{
    key: "dimensionValue",
    icon: Compare,
    serializer: () => ({}),
    deserializer: () => undefined,
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.values[0]]?.dimensions[ctx.values[1]];
      const dimLabel = getDimensionDisplayName(dim) || ctx.values[1];
      return dimLabel + ": " + ctx.values.slice(2).join(", ");
    },
  },
};
