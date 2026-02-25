import { getActiveMetricsViewNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";
import { createQuery } from "@tanstack/svelte-query";
import {
  getIdForContext,
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { MessageType } from "@rilldata/web-common/features/chat/core/types.ts";
import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
import { getLatestConversationQueryOptions } from "@rilldata/web-common/features/chat/core/utils.ts";

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view.
 * 1st level: metrics view context options
 * 2nd level: measures and dimensions options for each metrics view
 */
export function getMetricsViewPickerOptions(
  client: RuntimeClient,
): Readable<PickerItem[]> {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(client),
    queryClient,
  );

  const lastUsedMetricsViewStore = getLastUsedMetricsViewNameStore(client);
  const activeMetricsViewStore = getActiveMetricsViewNameStore(client);

  return derived(
    [metricsViewsQuery, lastUsedMetricsViewStore, activeMetricsViewStore],
    ([metricsViewsResp, lastUsedMetricsView, activeMetricsView]) => {
      const metricsViews = metricsViewsResp.data ?? [];
      return metricsViews.flatMap((mv) => {
        const mvName = mv.meta?.name?.name ?? "";
        const metricsViewSpec = mv.metricsView?.state?.validSpec ?? {};
        const mvContext = {
          type: InlineContextType.MetricsView,
          value: mvName,
          metricsView: mvName,
        } satisfies InlineContext;
        const mvPickerItem = {
          id: getIdForContext(mvContext),
          context: mvContext,
          currentlyActive: activeMetricsView === mvName,
          recentlyUsed: lastUsedMetricsView === mvName,
          hasChildren: true,
        } satisfies PickerItem;

        const measures = metricsViewSpec?.measures ?? [];
        const measurePickerItems = measures.map((m) => {
          const measureContext = {
            type: InlineContextType.Measure,
            value: m.name!,
            metricsView: mvName,
            measure: m.name!,
          } satisfies InlineContext;
          return {
            id: getIdForContext(measureContext),
            context: measureContext,
            parentId: mvPickerItem.id,
          } satisfies PickerItem;
        });

        const dimensions = metricsViewSpec?.dimensions ?? [];
        const dimensionPickerItems = dimensions.map((d) => {
          const dimensionContext = {
            type: InlineContextType.Dimension,
            value: d.name!,
            metricsView: mvName,
            dimension: d.name!,
          } satisfies InlineContext;
          return {
            id: getIdForContext(dimensionContext),
            context: dimensionContext,
            parentId: mvPickerItem.id,
          } satisfies PickerItem;
        });

        return [mvPickerItem, ...measurePickerItems, ...dimensionPickerItems];
      });
    },
  );
}

/**
 * Looks at the last conversation and returns the metrics view used in the last message or tool call.
 */
function getLastUsedMetricsViewNameStore(client: RuntimeClient) {
  const lastConversationQuery = createQuery(
    getLatestConversationQueryOptions(client),
    queryClient,
  );

  return derived(lastConversationQuery, (latestConversation) => {
    if (!latestConversation.data?.messages) return null;

    for (const message of latestConversation.data.messages) {
      if (message.type === MessageType.CALL) continue;
      const content = message.content?.[0];
      if (content?.toolCall?.input?.metrics_view) {
        return content.toolCall.input.metrics_view as string;
      }
    }

    return null;
  });
}
