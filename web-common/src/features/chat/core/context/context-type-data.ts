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

type ContextDataPerType<K extends keyof ContextRecord = keyof ContextRecord> = {
  key: K;
  icon: typeof ExploreIcon;
  serializer: (value: ContextRecord[K]) => RuntimeServiceCompleteBody;
  deserializer: (rawContext: Record<string, unknown>) => ContextRecord[K];
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
  },
  [ConversationContextType.Where]: <ContextDataPerType<"where">>{
    key: "where",
    icon: Filter,
    serializer: (where) => ({ where }),
    deserializer: (rawContext) =>
      mapResolverExpressionToV1Expression(rawContext.where as any),
  },
  [ConversationContextType.Measures]: <ContextDataPerType<"measures">>{
    key: "measures",
    icon: Compare,
    serializer: (measures) => ({ measures }),
    deserializer: (rawContext) =>
      (rawContext.measures as Array<unknown>)?.length
        ? rawContext.measures
        : undefined,
  },
  [ConversationContextType.Dimensions]: <ContextDataPerType<"dimensions">>{
    key: "dimensions",
    icon: Compare,
    serializer: (dimensions) => ({ dimensions }),
    deserializer: (rawContext) =>
      (rawContext.dimensions as Array<unknown>)?.length
        ? rawContext.dimensions
        : undefined,
  },
};
