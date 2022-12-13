import { goto } from "$app/navigation";
import type {
  V1PutFileAndReconcileResponse,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
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
      path: getFilePathFromNameAndType(tableName, EntityType.Table),
      blob: yaml,
      create: true,
      createOnly: true,
      strict: true,
    },
  });

  if (resp.errors.length) {
    return resp.errors;
  }
  goto(`/source/${tableName}`);
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  notifications.send({ message: `Created source ${tableName}` });
  return [];
}
