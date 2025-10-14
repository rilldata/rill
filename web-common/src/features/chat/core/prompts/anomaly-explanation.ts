import { ConversationContextType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { isExpressionEmpty } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function anomalyExplanation(
  instanceId: string,
  metricsViewName: string,
  filters: V1Expression,
) {
  if (!measureSelection.hasSelection()) return;

  const prompt = `Please explain what drives this data point. What dimensions have noticeably changed, as compared to other time windows?`;

  const context = [
    {
      type: ConversationContextType.MetricsView,
      value: metricsViewName,
    },
    ...measureSelection.getContexts(),
  ];
  if (!isExpressionEmpty(filters)) {
    context.push({
      type: ConversationContextType.Filters,
      // TODO: measure threshold filters
      value: convertExpressionToFilterParam(filters, []),
    });
  }

  sidebarActions.startChat(instanceId, prompt, context);
}
