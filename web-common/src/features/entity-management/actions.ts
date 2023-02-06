import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import type {
  V1DeleteFileAndReconcileResponse,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import type { ActiveEntity } from "@rilldata/web-local/lib/application-state-stores/app-store";
import {
  invalidateAfterReconcile,
  removeEntityQueries,
} from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import {
  getFilePathFromNameAndType,
  getLabel,
  getRouteFromName,
} from "./entity-mappers";
import { fileArtifactsStore } from "./file-artifacts-store";
import { getNextEntityName } from "./name-utils";
import type { EntityType } from "./types";

export async function renameFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  fromName: string,
  toName: string,
  type: EntityType,
  renameMutation: UseMutationResult<V1RenameFileAndReconcileResponse>
) {
  const resp = await renameMutation.mutateAsync({
    data: {
      instanceId,
      fromPath: getFilePathFromNameAndType(fromName, type),
      toPath: getFilePathFromNameAndType(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });

  removeEntityQueries(
    queryClient,
    instanceId,
    getFilePathFromNameAndType(fromName, type)
  );
  invalidateAfterReconcile(queryClient, instanceId, resp);
}

export async function deleteFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  name: string,
  type: EntityType,
  deleteMutation: UseMutationResult<V1DeleteFileAndReconcileResponse>,
  activeEntity: ActiveEntity,
  names: Array<string>,
  showNotification = true
) {
  const path = getFilePathFromNameAndType(name, type);
  try {
    const resp = await deleteMutation.mutateAsync({
      data: {
        instanceId,
        path,
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });
    }

    removeEntityQueries(queryClient, instanceId, path);

    invalidateAfterReconcile(queryClient, instanceId, resp);
    if (activeEntity?.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }
  } catch (err) {
    console.error(err);
  }
}
