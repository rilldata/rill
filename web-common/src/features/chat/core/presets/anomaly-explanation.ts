import { ConversationContextType } from "@rilldata/web-common/features/chat/core/types.ts";
import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime } from "luxon";

export function anomalyExplanation({
  metricsViewName,
  measure,
  hoveredTime,
  scrubStart,
  scrubEnd,
  timeGrain,
  zone,
}: {
  metricsViewName: string;
  measure: string;
  hoveredTime: Date;
  scrubStart: Date | null;
  scrubEnd: Date | null;
  timeGrain: V1TimeGrain;
  zone: string;
}) {
  const prompt = `Please explain what drives this data point. What dimensions have noticeably changed, as compared to other time windows?`;

  let timeRange = "";
  if (scrubStart && scrubEnd) {
    const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);
    timeRange = `${truncate(start, timeGrain, zone)} to ${truncate(end, timeGrain, zone)}`;
  } else {
    timeRange = truncate(hoveredTime, timeGrain, zone);
  }

  eventBus.emit("chat-intent", {
    prompt,
    context: [
      {
        type: ConversationContextType.MetricsView,
        value: metricsViewName,
      },
      {
        type: ConversationContextType.TimeRange,
        value: timeRange,
      },
      {
        type: ConversationContextType.Measures,
        value: measure,
      },
    ],
  });
}

function truncate(dt: Date, grain: V1TimeGrain, zone: string): string {
  return DateTime.fromJSDate(dt)
    .setZone(zone)
    .startOf(TIME_GRAIN[grain].label)
    .toJSDate()
    .toISOString();
}
