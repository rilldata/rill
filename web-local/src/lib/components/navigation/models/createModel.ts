import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  V1PutFileAndMigrateResponse,
} from "@rilldata/web-common/runtime-client";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getName } from "@rilldata/web-local/common/utils/incrementName";
import type { RuntimeState } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { commonEntitiesStore } from "@rilldata/web-local/lib/application-state-stores/common-store";
import { getFileFromName } from "@rilldata/web-local/lib/components/entity-mappers/mappers";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function createModel(
  runtimeState: RuntimeState,
  newModelName: string,
  createModelMutation: UseMutationResult<V1PutFileAndMigrateResponse>, // TODO: type
  sql = "",
  setAsActive = true
) {
  const resp = await createModelMutation.mutateAsync({
    data: {
      repoId: runtimeState.repoId,
      instanceId: runtimeState.instanceId,
      path: getFileFromName(newModelName, EntityType.Model),
      blob: sql,
      create: true,
      createOnly: true,
      strict: true,
    },
  });
  commonEntitiesStore.consolidateMigrateResponse(
    resp.affectedPaths,
    resp.errors
  );
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
    getRuntimeServiceListFilesQueryKey(runtimeState.repoId)
  );
}

export async function createModelFromSource(
  runtimeState: RuntimeState,
  modelNames: Array<string>,
  sourceName: string,
  createModelMutation: UseMutationResult<V1PutFileAndMigrateResponse>, // TODO: type
  setAsActive = true
): Promise<string> {
  const newModelName = getName(`${sourceName}_model`, modelNames);
  await createModel(
    runtimeState,
    newModelName,
    createModelMutation,
    `select * from ${sourceName}`,
    setAsActive
  );
  return newModelName;
}
