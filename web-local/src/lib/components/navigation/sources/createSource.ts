import { goto } from "$app/navigation";
import type {
  V1PutFileAndReconcileResponse,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import { EntityType } from "@rilldata/web-local/lib/temp/entity";
import { getFileFromName } from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient } from "@sveltestack/svelte-query";
import type { UseMutationResult } from "@sveltestack/svelte-query";
import { notifications } from "../../notifications";

export async function createSource(
  queryClient: QueryClient,
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
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  if (resp.errors.length) {
    // TODO: make sure to get the right error
    return resp.errors;
  }
  goto(`/source/${tableName}`);
  notifications.send({ message: `Created source ${tableName}` });
  return [];
}
