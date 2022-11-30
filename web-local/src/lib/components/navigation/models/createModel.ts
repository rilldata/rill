import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  V1PutFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getName } from "@rilldata/web-local/common/utils/incrementName";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";
import { getFileFromName } from "../../../util/entity-mappers";
import { notifications } from "../../notifications";

export async function createModel(
  instanceId: string,
  newModelName: string,
  createModelMutation: UseMutationResult<V1PutFileAndReconcileResponse>, // TODO: type
  sql = "",
  setAsActive = true
) {
  const resp = await createModelMutation.mutateAsync({
    data: {
      instanceId,
      path: getFileFromName(newModelName, EntityType.Model),
      blob: sql,
      create: true,
      createOnly: true,
      strict: true,
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  if (resp.errors?.length && sql !== "") {
    resp.errors.forEach((error) => {
      console.error(error);
    });
    throw new Error(resp.errors[0].filePath);
  }
  await dataModelerService.dispatch("addModel", [
    { name: newModelName, query: sql, asynchronous: true },
  ]);
  if (!setAsActive) return;
  goto(`/model/${newModelName}`);
  return queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(instanceId)
  );
}

export async function createModelFromSource(
  instanceId: string,
  modelNames: Array<string>,
  sourceName: string,
  createModelMutation: UseMutationResult<V1PutFileAndReconcileResponse>, // TODO: type
  setAsActive = true
): Promise<string> {
  const newModelName = getName(`${sourceName}_model`, modelNames);
  await createModel(
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
