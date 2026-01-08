import { getActiveMetricsViewNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { getValidMetricsViewsQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { derived, type Readable } from "svelte/store";
import type { InlineContextPickerParentOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { createQuery } from "@tanstack/svelte-query";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import {
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryOptions,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { MessageType } from "@rilldata/web-common/features/chat/core/types.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view.
 * 1st level: metrics view context options
 * 2nd level: measures and dimensions options for each metrics view
 */
export function getMetricsViewPickerOptions(): Readable<
  InlineContextPickerParentOption[]
> {
  const metricsViewsQuery = createQuery(
    getValidMetricsViewsQueryOptions(),
    queryClient,
  );

  const lastUsedMetricsViewStore = getLastUsedMetricsViewNameStore();
  const activeMetricsViewStore = getActiveMetricsViewNameStore();

  return derived(
    [metricsViewsQuery, lastUsedMetricsViewStore, activeMetricsViewStore],
    ([metricsViewsResp, lastUsedMetricsView, activeMetricsView]) => {
      const metricsViews = metricsViewsResp.data ?? [];
      return metricsViews.map((mv) => {
        const mvName = mv.meta?.name?.name ?? "";
        const metricsViewSpec = mv.metricsView?.state?.validSpec ?? {};
        const mvDisplayName = metricsViewSpec?.displayName || mvName;

        const measures: InlineContext[] =
          metricsViewSpec?.measures?.map((m) => ({
            type: InlineContextType.Measure,
            label: getMeasureDisplayName(m),
            key: `${InlineContextType.Measure}_${m.name!}`,
            metricsView: mvName,
            measure: m.name!,
          })) ?? [];

        const dimensions: InlineContext[] =
          metricsViewSpec?.dimensions?.map((d) => ({
            type: InlineContextType.Dimension,
            label: getDimensionDisplayName(d),
            key: `${InlineContextType.Dimension}_${d.name!}`,
            metricsView: mvName,
            dimension: d.name!,
          })) ?? [];

        return {
          context: {
            type: InlineContextType.MetricsView,
            metricsView: mvName,
            label: mvDisplayName,
            key: `${InlineContextType.MetricsView}_${mvName}`,
          },
          children: measures.concat(dimensions),
          currentlyActive: activeMetricsView === mvName,
          recentlyUsed: lastUsedMetricsView === mvName,
        } satisfies InlineContextPickerParentOption;
      });
    },
  );
}

/**
 * Looks at the last conversation and returns the metrics view used in the last message or tool call.
 */
export function getLastUsedMetricsViewNameStore() {
  const lastConversationQuery = createQuery(
    getLatestConversationQueryOptions(),
    queryClient,
  );

  return derived(lastConversationQuery, (latestConversation) => {
    if (!latestConversation.data?.messages) return null;

    for (const message of latestConversation.data.messages) {
      const metricsView = getMetricsViewInMessage(message);
      if (metricsView) return metricsView;
    }

    return null;
  });
}

/**
 * Returns the last updated conversation ID.
 */
function getLatestConversationQueryOptions() {
  const listConversationsQueryOptions = derived(runtime, ({ instanceId }) =>
    getRuntimeServiceListConversationsQueryOptions(instanceId, {
      // Filter to only show Rill client conversations, excluding MCP conversations
      userAgentPattern: "rill%",
    }),
  );
  const lastConversationId = derived(
    createQuery(listConversationsQueryOptions, queryClient),
    (conversationsResp) => {
      return conversationsResp?.data?.conversations?.[0]?.id;
    },
  );

  return derived(
    [lastConversationId, runtime],
    ([lastConversationId, { instanceId }]) => {
      return getRuntimeServiceGetConversationQueryOptions(
        instanceId,
        lastConversationId ?? "",
        {
          query: {
            enabled: !!lastConversationId,
          },
        },
      );
    },
  );
}

function getMetricsViewInMessage(message: V1Message) {
  if (message.type !== MessageType.CALL) return null;
  const content = message.content?.[0];
  return (content?.toolCall?.input?.metrics_view as string) ?? null;
}
