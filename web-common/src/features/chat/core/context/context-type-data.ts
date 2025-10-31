import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  type RuntimeServiceCompleteBody,
  type V1Expression,
  V1TimeGrain,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { derived, readable, type Readable } from "svelte/store";

export enum ConversationContextType {
  Explore = "explore",
  TimeRange = "timeRange",
  Where = "where",
  Measures = "measures",
  Dimensions = "dimensions",
}
export const FILTER_CONTEXT_TYPES = [
  ConversationContextType.TimeRange,
  ConversationContextType.Where,
];

export type ContextRecord = {
  explore?: string;
  dimensions?: string[];
  measures?: string[];
  where?: V1Expression;
  timeRange?: V1TimeRange;
};

export type ConversationContextEntry<
  K extends keyof ContextRecord = keyof ContextRecord,
> = {
  type: ConversationContextType;
  value: ContextRecord[K];
};

type ContextDataPerType<K extends keyof ContextRecord = keyof ContextRecord> = {
  key: K;
  icon: typeof ExploreIcon;
  serializer: (value: ContextRecord[K]) => RuntimeServiceCompleteBody;
  deserializer: (rawContext: Record<string, unknown>) => ContextRecord[K];
  formatter: (
    value: ContextRecord[K],
    record: ContextRecord,
    instanceId: string,
  ) => Readable<string>;
};

export const ContextTypeData: Record<
  ConversationContextType,
  ContextDataPerType
> = {
  [ConversationContextType.Explore]: <ContextDataPerType<"explore">>{
    key: "explore",
    icon: ExploreIcon,
    serializer: (explore) => ({ explore }),
    deserializer: (rawContext) => rawContext.explore,
    formatter: (exploreName, _, instanceId) =>
      derived(
        useExploreValidSpec(instanceId, exploreName ?? ""),
        (metricsViewResp) =>
          metricsViewResp.data?.explore?.displayName || exploreName,
      ),
  },
  [ConversationContextType.TimeRange]: <ContextDataPerType<"timeRange">>{
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
    formatter: (timeRange) => {
      if (!timeRange?.start || !timeRange?.end) return readable("");

      return readable(
        prettyFormatTimeRange(
          Interval.fromDateTimes(
            DateTime.fromISO(timeRange.start),
            DateTime.fromISO(timeRange.end),
          ),
          V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        ),
      );
    },
  },
  [ConversationContextType.Where]: <ContextDataPerType<"where">>{
    key: "where",
    icon: Filter,
    serializer: (where) => ({ where }),
    deserializer: (rawContext) =>
      mapResolverExpressionToV1Expression(rawContext.where as any),
    formatter: (where) =>
      readable(where ? convertExpressionToFilterParam(where, []) : ""),
  },
  [ConversationContextType.Measures]: <ContextDataPerType<"measures">>{
    key: "measures",
    icon: Compare,
    serializer: (measures) => ({ measures }),
    deserializer: (rawContext) =>
      (rawContext.measures as Array<unknown>)?.length
        ? rawContext.measures
        : undefined,
    formatter: (measureNames, contextRecord, instanceId) =>
      derived(
        useExploreValidSpec(
          instanceId,
          contextRecord[ConversationContextType.Explore] ?? "",
        ),
        (metricsViewResp) => {
          const measureDisplayNames = measureNames?.map(
            (measureName) =>
              getMeasureDisplayName(
                metricsViewResp.data?.metricsView?.measures?.find(
                  (m) => m.name === measureName,
                ),
              ) ?? measureName,
          );
          return measureDisplayNames?.join(", ") ?? "";
        },
      ),
  },
  [ConversationContextType.Dimensions]: <ContextDataPerType<"dimensions">>{
    key: "dimensions",
    icon: Compare,
    serializer: (dimensions) => ({ dimensions }),
    deserializer: (rawContext) =>
      (rawContext.dimensions as Array<unknown>)?.length
        ? rawContext.dimensions
        : undefined,
    formatter: (dimensionNames, contextRecord, instanceId) =>
      derived(
        useExploreValidSpec(
          instanceId,
          contextRecord[ConversationContextType.Explore] ?? "",
        ),
        (metricsViewResp) => {
          const dimensionDisplayNames = dimensionNames?.map(
            (dimensionName) =>
              getDimensionDisplayName(
                metricsViewResp.data?.metricsView?.dimensions?.find(
                  (m) => m.name === dimensionName,
                ),
              ) ?? dimensionName,
          );
          return dimensionDisplayNames?.join(", ") ?? "";
        },
      ),
  },
};
