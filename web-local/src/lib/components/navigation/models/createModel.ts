import { goto } from "$app/navigation";
import type { V1PutFileAndReconcileResponse } from "@rilldata/web-common/runtime-client";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getName } from "@rilldata/web-local/common/utils/incrementName";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import { getFilePathFromNameAndType } from "../../../util/entity-mappers";
import { notifications } from "../../notifications";

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
      create: true,
      createOnly: true,
      strict: true,
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  goto(`/model/${newModelName}`);
  invalidateAfterReconcile(queryClient, instanceId, resp);
  if (resp.errors?.length && sql !== "") {
    resp.errors.forEach((error) => {
      console.error(error);
    });
    throw new Error(resp.errors[0].filePath);
  }
  if (!setAsActive) return;
}

export async function createModelFromSource(
  queryClient: QueryClient,
  instanceId: string,
  modelNames: Array<string>,
  sourceName: string,
  createModelMutation: UseMutationResult<V1PutFileAndReconcileResponse>, // TODO: type
  setAsActive = true
): Promise<string> {
  const newModelName = getName(`${sourceName}_model`, modelNames);
  await createModel(
    queryClient,
    instanceId,
    newModelName,
    createModelMutation,
    `select * from ${sourceName}`,
    setAsActive
  );
  notifications.send({
    message: `Queried ${sourceName} in workspace`,
  });
  return newModelName;
}
