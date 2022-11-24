import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  V1MigrationError,
  V1PutFileAndMigrateResponse,
} from "@rilldata/web-common/runtime-client";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import type { RuntimeState } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { commonEntitiesStore } from "@rilldata/web-local/lib/application-state-stores/common-store";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function createSource(
  runtimeState: RuntimeState,
  tableName: string,
  yaml: string,
  createSourceMutation: UseMutationResult<V1PutFileAndMigrateResponse>
): Promise<V1MigrationError[]> {
  const resp = await createSourceMutation.mutateAsync({
    data: {
      repoId: runtimeState.repoId,
      instanceId: runtimeState.instanceId,
      path: `sources/${tableName}.yaml`,
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
    getRuntimeServiceListFilesQueryKey(runtimeState.repoId)
  );
  return [];
}
