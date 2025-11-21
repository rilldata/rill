import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
import {
  getDimensionDisplayName,
  getExploreDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { prettyFormatV1TimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type RuntimeServiceCompleteBody,
  type V1Expression,
  type V1Resource,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export enum ChatContextEntryType {
  Explore = "explore",
  MetricsView = "metrics_view",
  TimeRange = "time_range",
  Where = "where",
  Measures = "measures",
  Dimensions = "dimensions",
  DimensionValues = "dimension_values",
}

export type ChatContextEntry = {
  type: ChatContextEntryType;
  label: string;
  value: string;
  subValue: string | null;
};

export type ContextRecord = {
  explore?: string;
  dimensions?: string[];
  measures?: string[];
  where?: V1Expression;
  timeRange?: V1TimeRange;
  dimensionValue?: string;
};

export type ContextMetadata = {
  validExplores: Record<string, V1Resource>;
  measures: Record<string, MetricsViewSpecMeasure>;
  dimensions: Record<string, MetricsViewSpecDimension>;
};
type ContextDataPerType<K extends keyof ContextRecord = keyof ContextRecord> = {
  key: K;
  icon: typeof ExploreIcon;
  serializer: (value: ContextRecord[K]) => RuntimeServiceCompleteBody;
  deserializer: (rawContext: Record<string, unknown>) => ContextRecord[K];
  getLabel: (ctx: ChatContextEntry, meta: ContextMetadata) => string;
};

export const ContextTypeData: Partial<
  Record<ChatContextEntryType, ContextDataPerType>
> = {
  [ChatContextEntryType.Explore]: <ContextDataPerType<"explore">>{
    key: "explore",
    icon: ExploreIcon,
    serializer: (explore) => ({ explore }),
    deserializer: (rawContext) => rawContext.explore,
    getLabel: (ctx, meta) =>
      getExploreDisplayName(meta.validExplores[ctx.value]) || ctx.value,
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
      const [start, end] = ctx.value.split(" to ");
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
    getLabel: (ctx) => ctx.value,
  },

  [ChatContextEntryType.Measures]: <ContextDataPerType<"measures">>{
    key: "measures",
    icon: Compare,
    serializer: (measures) => ({ measures }),
    deserializer: (rawContext) =>
      (rawContext.measures as Array<unknown>)?.length
        ? rawContext.measures
        : undefined,
    getLabel: (ctx, meta) =>
      getMeasureDisplayName(meta.measures[ctx.value]) || ctx.value,
  },

  [ChatContextEntryType.Dimensions]: <ContextDataPerType<"dimensions">>{
    key: "dimensions",
    icon: Compare,
    serializer: (dimensions) => ({ dimensions }),
    deserializer: (rawContext) =>
      (rawContext.dimensions as Array<unknown>)?.length
        ? rawContext.dimensions
        : undefined,
    getLabel: (ctx, meta) =>
      getDimensionDisplayName(meta.dimensions[ctx.value]) || ctx.value,
  },

  [ChatContextEntryType.DimensionValues]: <
    ContextDataPerType<"dimensionValue">
  >{
    key: "dimensionValue",
    icon: Compare,
    serializer: () => ({}),
    deserializer: () => undefined,
    getLabel: (ctx, meta) =>
      (getDimensionDisplayName(meta.dimensions[ctx.value]) || ctx.value) +
      ": " +
      (ctx.subValue ?? ""),
  },
};
