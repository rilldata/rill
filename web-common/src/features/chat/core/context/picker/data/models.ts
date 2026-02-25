import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { getQueryServiceTableColumnsQueryOptions } from "@rilldata/web-common/runtime-client/v2/gen";
import {
  getIdForContext,
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { getActiveResourceStore } from "@rilldata/web-common/features/entity-management/nav-utils.ts";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, type Readable } from "svelte/store";
import { createQuery } from "@tanstack/svelte-query";
import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

/**
 * Creates a store that contains a 2-level list of sources/model resources.
 * 1st level: section for sources/models.
 * 2nd level: all the columns in the source/model resource.
 * NOTE: this only lists resources that are parsed as sources/models. Any parse errors will exclude the file.
 */
export function getModelsPickerOptions(
  client: RuntimeClient,
  uiState: ContextPickerUIState,
): Readable<PickerItem[]> {
  const modelResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(client, ResourceKind.Model),
    queryClient,
  );
  const activeResourceStore = getActiveResourceStore();

  return derived(
    [modelResourcesQuery, activeResourceStore],
    ([modelResourcesResp, activeResource], set) => {
      const models = modelResourcesResp.data ?? [];
      const modelPickerItems: PickerItem[] = [];
      const modelQueryOptions: ReturnType<
        typeof getModelColumnsQueryOptions
      >[] = [];

      models.forEach((res) => {
        const modelName = res.meta?.name?.name ?? "";

        const currentlyActive =
          activeResource?.kind === ResourceKind.Model &&
          activeResource?.name === modelName;
        const modelContext = {
          type: InlineContextType.Model,
          model: modelName,
          value: modelName,
        } satisfies InlineContext;
        const modelPickerItem = {
          id: getIdForContext(modelContext),
          context: modelContext,
          currentlyActive,
          hasChildren: true,
        } satisfies PickerItem;

        const childrenQueryOptions = getModelColumnsQueryOptions(
          client,
          res,
          modelPickerItem,
          uiState.getExpandedStore(modelPickerItem.id),
        );

        modelPickerItems.push(modelPickerItem);
        modelQueryOptions.push(childrenQueryOptions);
      });

      const allPickerOptionsStore = derived(
        modelQueryOptions.map((o) => createQuery(o, queryClient)),
        (modelQueryResults) => {
          return modelQueryResults.flatMap((res, index) => [
            {
              ...modelPickerItems[index],
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

function getModelColumnsQueryOptions(
  client: RuntimeClient,
  modelRes: V1Resource | undefined,
  modelPickerItem: PickerItem,
  enabledStore: Readable<boolean>,
) {
  const connector = modelRes?.model?.spec?.outputConnector ?? "";
  const table = modelRes?.model?.state?.resultTable ?? "";
  return derived(enabledStore, (enabled) =>
    getQueryServiceTableColumnsQueryOptions(
      client,
      {
        tableName: table,
        connector,
      },
      {
        query: {
          enabled: enabled && Boolean(table),
          select: (data): PickerItem[] => {
            return (
              data.profileColumns?.map((col) => {
                const context = {
                  type: InlineContextType.Column,
                  value: col.name!,
                  column: col.name,
                  columnType: col.type,
                  model: modelPickerItem.context.model,
                } satisfies InlineContext;

                return {
                  id: getIdForContext(context),
                  context,
                  parentId: modelPickerItem.id,
                } satisfies PickerItem;
              }) ?? []
            );
          },
        },
      },
    ),
  );
}
