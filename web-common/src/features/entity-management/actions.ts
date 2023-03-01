import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import { currentHref } from "@rilldata/web-common/layout/navigation/stores";
import type {
  V1DeleteFileAndReconcileResponse,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { getHttpRequestQueueForHost } from "@rilldata/web-common/runtime-client/http-client";
import type { ActiveEntity } from "@rilldata/web-local/lib/application-state-stores/app-store";
import {
  invalidateAfterReconcile,
  removeEntityQueries,
} from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import { get } from "svelte/store";
import { runtime } from "../../runtime-client/runtime-store";
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
  fromName: string,
  toName: string,
  type: EntityType,
  renameMutation: UseMutationResult<V1RenameFileAndReconcileResponse>
) {
  const resp = await renameMutation.mutateAsync({
    data: {
      instanceId: get(runtime).instanceId,
      fromPath: getFilePathFromNameAndType(fromName, type),
      toPath: getFilePathFromNameAndType(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  const httpRequestQueue = getHttpRequestQueueForHost(get(runtime).host);
  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });

  removeEntityQueries(queryClient, getFilePathFromNameAndType(fromName, type));
  invalidateAfterReconcile(queryClient, resp);
}

export async function deleteFileArtifact(
  queryClient: QueryClient,
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
        instanceId: get(runtime).instanceId,
        path,
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    const httpRequestQueue = getHttpRequestQueueForHost(get(runtime).host);
    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });
    }

    removeEntityQueries(queryClient, path);

    invalidateAfterReconcile(queryClient, resp);
    if (activeEntity?.name === name) {
      const route = getRouteFromName(getNextEntityName(names, name), type);
      /** set the href store so the menu selection has an immediate visual update. */
      currentHref.set(route);
      goto(route);
    }
  } catch (err) {
    console.error(err);
  }
}
