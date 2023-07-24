import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { runtimeServicePutFileAndReconcile } from "../../runtime-client";
import { invalidateAfterReconcile } from "../../runtime-client/invalidation";
import { runtime } from "../../runtime-client/runtime-store";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { fileArtifactsStore } from "../entity-management/file-artifacts-store";
import { EntityType } from "../entity-management/types";

export async function saveAndRefresh(
  queryClient: QueryClient,
  tableName: string,
  yaml: string
) {
  const instanceId = get(runtime).instanceId;
  const resp = await runtimeServicePutFileAndReconcile({
    instanceId,
    path: getFilePathFromNameAndType(tableName, EntityType.Table),
    blob: yaml,
    create: false,
    createOnly: false,
    dry: false,
    strict: false,
  });

  invalidateAfterReconcile(queryClient, instanceId, resp);

  // handle errors
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
}
