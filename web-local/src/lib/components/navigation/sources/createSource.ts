import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  V1PutFileAndReconcileResponse,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { commonEntitiesStore } from "@rilldata/web-local/lib/application-state-stores/common-store";
import { getFileFromName } from "@rilldata/web-local/lib/components/entity-mappers/mappers";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function createSource(
  instanceId: string,
  tableName: string,
  yaml: string,
  createSourceMutation: UseMutationResult<V1PutFileAndReconcileResponse>
): Promise<V1ReconcileError[]> {
  const resp = await createSourceMutation.mutateAsync({
    data: {
      instanceId,
      path: getFileFromName(tableName, EntityType.Table),
      blob: yaml,
      create: true,
      createOnly: true,
      strict: true,
    },
  });
  commonEntitiesStore.consolidateMigrateResponse(
    resp.affectedPaths,
    resp.errors
  );
  if (resp.errors.length) {
    // TODO: make sure to get the right error
    return resp.errors;
  }
  await dataModelerService.dispatch("addOrSyncTableFromDB", [tableName, true]);
  goto(`/source/${tableName}`);
  await queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(instanceId)
  );
  return [];
}
