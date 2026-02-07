import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
import { derived, get, type Readable } from "svelte/store";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { getCanvasNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { createQuery } from "@tanstack/svelte-query";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { getQueryServiceResolveCanvasQueryOptions } from "@rilldata/web-common/runtime-client";
import {
  getIdForContext,
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import { getLatestConversationQueryOptions } from "@rilldata/web-common/features/chat/core/utils.ts";
import { MessageType } from "@rilldata/web-common/features/chat/core/types.ts";

export function getCanvasesPickerOptions(
  uiState: ContextPickerUIState,
): Readable<PickerItem[]> {
  const canvasResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(ResourceKind.Canvas, (res) =>
      Boolean(res.canvas?.state?.validSpec),
    ),
    queryClient,
  );
  const lastUsedCanvasNameStore = getLastUsedCanvasNameStore();
  const activeCanvasNameStore = getCanvasNameStore();
  const instanceId = get(runtime).instanceId;

  return derived(
    [canvasResourcesQuery, lastUsedCanvasNameStore, activeCanvasNameStore],
    ([canvasResourcesQueryResp, lastUsedCanvasName, activeCanvasName], set) => {
      const canvases = canvasResourcesQueryResp.data ?? [];
      const canvasPickerItems: PickerItem[] = [];
      const canvasQueryOptions: ReturnType<
        typeof getCanvasComponentsQueryOptions
      >[] = [];

      canvases.forEach((res) => {
        const canvasName = res.meta?.name?.name ?? "";

        const canvasContext = {
          type: InlineContextType.Canvas,
          value: canvasName,
          canvas: canvasName,
        } satisfies InlineContext;
        const canvasPickerItem = {
          id: getIdForContext(canvasContext),
          context: canvasContext,
          currentlyActive: activeCanvasName === canvasName,
          recentlyUsed: lastUsedCanvasName === canvasName,
          hasChildren: true,
        } satisfies PickerItem;

        const childrenQueryOptions = getCanvasComponentsQueryOptions(
          instanceId,
          canvasPickerItem,
          uiState.getExpandedStore(canvasPickerItem.id),
        );

        canvasPickerItems.push(canvasPickerItem);
        canvasQueryOptions.push(childrenQueryOptions);
      });

      const allPickerOptionsStore = derived(
        canvasQueryOptions.map((o) => createQuery(o, queryClient)),
        (canvasQueryResults) => {
          return canvasQueryResults.flatMap((res, index) => [
            {
              ...canvasPickerItems[index],
              childrenLoading: res.isLoading,
            },
            ...(res.data ?? []),
          ]);
        },
      );

      return allPickerOptionsStore.subscribe(set);
    },
  );
}

function getCanvasComponentsQueryOptions(
  instanceId: string,
  canvasPickerItem: PickerItem,
  enabledStore: Readable<boolean>,
) {
  const canvas = canvasPickerItem.context.canvas!;
  return derived(enabledStore, (enabled) =>
    getQueryServiceResolveCanvasQueryOptions(
      instanceId,
      canvas,
      {},
      {
        query: {
          select: (data): PickerItem[] => {
            const componentItems = Object.entries(data.resolvedComponents ?? [])
              .map(([name, res]) => {
                const componentSpec = res.component?.state?.validSpec;
                if (!componentSpec) return null;

                const componentContext = {
                  type: InlineContextType.CanvasComponent,
                  value: name,
                  canvas,
                  canvasComponent: name,
                } satisfies InlineContext;

                return {
                  id: getIdForContext(componentContext),
                  context: componentContext,
                  parentId: canvasPickerItem.id,
                } satisfies PickerItem;
              })
              .filter(Boolean) as PickerItem[];

            // Since we are converting object to key-value array ensure order is consistent.
            // While JS maintains inserted order, golang (backend) has random order.
            componentItems.sort((a, b) => (a.id > b.id ? 1 : -1));
            return componentItems;
          },
          enabled,
        },
      },
    ),
  );
}

/**
 * Looks at the last conversation and returns the canvas used in the last message or tool call.
 */
function getLastUsedCanvasNameStore() {
  const lastConversationQuery = createQuery(
    getLatestConversationQueryOptions(),
    queryClient,
  );

  return derived(lastConversationQuery, (latestConversation) => {
    if (!latestConversation.data?.messages) return null;

    for (const message of latestConversation.data.messages) {
      if (message.type === MessageType.CALL) continue;
      const content = message.content?.[0];
      if (content?.toolCall?.input?.canvas) {
        return content.toolCall.input.canvas as string;
      }
    }

    return null;
  });
}
