import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { derived, readable, type Readable } from "svelte/store";

export enum ConversationContextType {
  MetricsView,
  TimeRange,
  Filters,
  Measure,
}

export type ConversationContextEntry = {
  type: ConversationContextType;
  value: string;
};

type ContextDataPerType = {
  prompt: string;
  icon: typeof MetricsViewIcon;
  formatter: (
    value: string,
    record: ContextRecord,
    instanceId: string,
  ) => Readable<string>;
};

export type ContextRecord = Partial<Record<ConversationContextType, string>>;

export const ContextTypeData: Record<
  ConversationContextType,
  ContextDataPerType
> = {
  [ConversationContextType.MetricsView]: {
    prompt: `Skip "list_metrics_views" tool call and use this metrics view instead`,
    icon: MetricsViewIcon,
    formatter: (metricsViewName, _, instanceId) =>
      derived(
        useMetricsView(instanceId, metricsViewName),
        (metricsViewResp) =>
          metricsViewResp.data?.metricsView?.state?.validSpec?.displayName ||
          metricsViewName,
      ),
  },
  [ConversationContextType.TimeRange]: {
    prompt: `Skip "query_metrics_view_summary" tool call and use this time range instead`,
    icon: Calendar,
    formatter: (range) => {
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
  [ConversationContextType.Filters]: {
    prompt: "Filters",
    icon: Filter,
    formatter: (filter) => readable(filter),
  },
  [ConversationContextType.Measure]: {
    prompt: "Measure",
    icon: Compare,
    formatter: (measureName, contextRecord, instanceId) =>
      derived(
        useMetricsView(
          instanceId,
          contextRecord[ConversationContextType.MetricsView] ?? "",
        ),
        (metricsViewResp) =>
          getMeasureDisplayName(
            metricsViewResp.data?.metricsView?.state?.validSpec?.measures?.find(
              (m) => m.name === measureName,
            ),
          ) || measureName,
      ),
  },
};

export function extractContextEntry(
  promptLine: string,
): ConversationContextEntry | null {
  for (const type in ContextTypeData) {
    const contextData = ContextTypeData[type];
    if (!promptLine.startsWith(contextData.prompt)) continue;

    const value = promptLine
      .replace(contextData.prompt, "")
      .trim()
      .replace(/^:\s*"/, "")
      .replace(/"$/, "");

    return {
      type: Number(type),
      value,
    };
  }

  return null;
}
