import { ConversationContextType } from "@rilldata/web-common/features/chat/core/types.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

export function anomalyExplanation(
  metricsViewName: string,
  timestamp: string,
  measure: string,
) {
  const prompt = `Please explain what drives this data point. What dimensions have noticeably changed, as compared to other time windows?`;

  eventBus.emit("chat-intent", {
    prompt,
    context: [
      {
        type: ConversationContextType.MetricsView,
        value: metricsViewName,
      },
      {
        type: ConversationContextType.TimeRange,
        value: timestamp,
      },
      {
        type: ConversationContextType.Measures,
        value: measure,
      },
    ],
  });
}
