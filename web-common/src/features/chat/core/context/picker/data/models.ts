import {
  getQueryServiceTableColumnsQueryOptions,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { InlineContextPickerParentOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { getActiveFileArtifactStore } from "@rilldata/web-common/features/entity-management/nav-utils.ts";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, type Readable } from "svelte/store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { createQuery } from "@tanstack/svelte-query";
import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";

/**
 * Creates a store that contains a 2-level list of sources/model resources.
 * 1st level: section for sources/models.
 * 2nd level: all the columns in the source/model resource.
 * NOTE: this only lists resources that are parsed as sources/models. Any parse errors will exclude the file.
 */
export function getModelsPickerOptions(
  uiState: ContextPickerUIState,
): Readable<InlineContextPickerParentOption[]> {
  const modelResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(ResourceKind.Model),
    queryClient,
  );
  const activeFileArtifactStore = getActiveFileArtifactStore();

  return derived(
    [runtime, modelResourcesQuery, activeFileArtifactStore],
    ([{ instanceId }, modelResourcesResp, activeFileArtifact]) => {
      const models = modelResourcesResp.data ?? [];
      return models.map((res) => {
        const modelName = res.meta?.name?.name ?? "";

        const key = `${InlineContextType.Model}_${modelName}`;
        const currentlyActive =
          activeFileArtifact.resource?.kind === ResourceKind.Model &&
          activeFileArtifact.resource?.name === modelName;

        return {
          context: {
            type: InlineContextType.Model,
            model: modelName,
            key,
            label: modelName,
          },
          childrenQueryOptions: getModelColumnsQueryOptions(
            instanceId,
            res,
            modelName,
            uiState.getExpandedStore(key),
          ),
          currentlyActive,
        } satisfies InlineContextPickerParentOption;
      });
    },
  );
}

function getModelColumnsQueryOptions(
  instanceId: string,
  modelRes: V1Resource | undefined,
  modelName: string,
  enabledStore: Readable<boolean>,
) {
  const connector = modelRes?.model?.spec?.outputConnector ?? "";
  const table = modelRes?.model?.state?.resultTable ?? "";
  return derived(enabledStore, (enabled) =>
    getQueryServiceTableColumnsQueryOptions(
      instanceId,
      table,
      {
        connector,
      },
      {
        query: {
          enabled: enabled && Boolean(table),
          select: (data) =>
            data.profileColumns?.map(
              (col) =>
                ({
                  type: InlineContextType.Column,
                  label: col.name,
                  key: `${InlineContextType.Column}_${col.name!}`,
                  column: col.name,
                  columnType: col.type,
                  model: modelName,
                }) satisfies InlineContext,
            ) ?? [],
        },
      },
    ),
  );
}
