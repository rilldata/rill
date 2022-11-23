import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  V1PutFileAndMigrateResponse,
} from "@rilldata/web-common/runtime-client";
import { getName } from "@rilldata/web-local/common/utils/incrementName";
import type { RuntimeState } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function createModel(
  runtimeState: RuntimeState,
  newModelName: string,
  createModelMutation: UseMutationResult<V1PutFileAndMigrateResponse>, // TODO: type
  sql = "",
  setAsActive = true
) {
  const res = await createModelMutation.mutateAsync({
    data: {
      repoId: runtimeState.repoId,
      instanceId: runtimeState.instanceId,
      path: `models/${newModelName}.sql`,
      blob: sql,
      create: true,
      createOnly: true,
      strict: true,
    },
  });
  if (res.errors?.length && sql !== "") {
    res.errors.forEach((error) => {
      console.error(error);
    });
    throw new Error(res.errors[0].filePath);
  }
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
