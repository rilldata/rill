import { goto } from "$app/navigation";
import type {
  V1DeleteFileAndReconcileResponse,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getNextEntityName } from "@rilldata/web-local/common/utils/getNextEntityId";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { notifications } from "@rilldata/web-local/lib/components/notifications";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import {
  getFileFromName,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient } from "@sveltestack/svelte-query";
import type { UseMutationResult } from "@sveltestack/svelte-query";

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
      fromPath: getFileFromName(fromName, type),
      toPath: getFileFromName(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  goto(getRouteFromName(toName, type), {
    replaceState: true,
  });
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
  return invalidateAfterReconcile(queryClient, instanceId, resp);
}

export async function deleteFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  name: string,
  type: EntityType,
  deleteMutation: UseMutationResult<V1DeleteFileAndReconcileResponse>,
  activeEntity: ActiveEntity,
  names: Array<string>
) {
  try {
    const resp = await deleteMutation.mutateAsync({
      data: {
        instanceId,
        path: getFileFromName(name, type),
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
    if (activeEntity?.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }

    notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });

    return invalidateAfterReconcile(queryClient, instanceId, resp);
  } catch (err) {
    console.error(err);
  }
}
