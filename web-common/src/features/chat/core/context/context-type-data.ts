import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  V1TimeGrain,
  type V1CompletionMessageContext,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { derived, readable, type Readable } from "svelte/store";

export enum ConversationContextType {
  Explore,
  TimeRange,
  Filters,
  Measure,
}

export type ConversationContextEntry<
  K extends keyof V1CompletionMessageContext = keyof V1CompletionMessageContext,
> = {
  type: ConversationContextType;
  value: V1CompletionMessageContext[K];
};

type ContextDataPerType<
  K extends keyof V1CompletionMessageContext = keyof V1CompletionMessageContext,
> = {
  key: K;
  icon: typeof ExploreIcon;
  formatter: (
    value: V1CompletionMessageContext[K],
    record: ContextRecord,
    instanceId: string,
  ) => Readable<string>;
};

export type ContextRecord = Partial<Record<ConversationContextType, string>>;

export const ContextTypeData: Record<
  ConversationContextType,
  ContextDataPerType
> = {
  [ConversationContextType.Explore]: <ContextDataPerType<"explore">>{
    key: "explore",
    icon: ExploreIcon,
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
    formatter: (range) => {
      if (!range) return readable("");

      const times = range.split(/\s*to\s*/);
      const start = times[0];
      const end = times[1] ?? start;
      return readable(
        prettyFormatTimeRange(
          Interval.fromDateTimes(
            DateTime.fromISO(start),
            DateTime.fromISO(end),
          ),
          V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        ),
      );
    },
  },
  [ConversationContextType.Filters]: <ContextDataPerType<"filters">>{
    key: "filters",
    icon: Filter,
    formatter: (filter) => readable(filter),
  },
  [ConversationContextType.Measure]: <ContextDataPerType<"measures">>{
    key: "measures",
    icon: Compare,
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
};

export const ContextKeyToTypeMap = Object.fromEntries(
  Object.entries(ContextTypeData).map(([type, value]) => [
    value.key,
    Number(type) as ConversationContextType,
  ]),
);
