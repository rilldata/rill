import { goto } from "$app/navigation";
import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import type { V1PutFileAndReconcileResponse } from "@rilldata/web-common/runtime-client";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";

export async function createModel(
  queryClient: QueryClient,
  instanceId: string,
  newModelName: string,
  createModelMutation: UseMutationResult<V1PutFileAndReconcileResponse>, // TODO: type
  sql = "",
  setAsActive = true
) {
  const resp = await createModelMutation.mutateAsync({
    data: {
      instanceId,
      path: getFilePathFromNameAndType(newModelName, EntityType.Model),
      blob: sql,
      syncFromUrl: true,
      createOnly: true,
      strict: true,
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  goto(`/model/${newModelName}?focus`);

  invalidateAfterReconcile(queryClient, instanceId, resp);
  if (resp.errors?.length && sql !== "") {
    resp.errors.forEach((error) => {
      console.error(error);
    });
    throw new Error(resp.errors[0].filePath);
  }

  if (!setAsActive) return;
}
