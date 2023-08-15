import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import type {
  V1PutFileAndReconcileResponse,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import { invalidateAfterReconcile } from "@rilldata/web-common/runtime-client/invalidation";
import type {
  CreateBaseMutationResult,
  QueryClient,
} from "@tanstack/svelte-query";

export async function createSource(
  queryClient: QueryClient,
  instanceId: string,
  tableName: string,
  yaml: string,
  createSourceMutation: CreateBaseMutationResult<V1PutFileAndReconcileResponse>
): Promise<V1ReconcileError[]> {
  const resp = await createSourceMutation.mutateAsync({
    data: {
      instanceId,
      path: getFilePathFromNameAndType(tableName, EntityType.Table),
      blob: yaml,
      // create source is used to upload and replace.
      // so we cannot send createOnly=true until we refactor it to use refresh source
      createOnly: false,
      strict: true,
    },
  });

  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  return resp.errors;
}
