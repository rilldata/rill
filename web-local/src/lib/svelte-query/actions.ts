import { goto } from "$app/navigation";
import type {
  V1DeleteFileAndReconcileResponse,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getNextEntityName } from "@rilldata/web-local/common/utils/getNextEntityId";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import {
  getFileFromName,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-local/lib/util/entity-mappers";
import notifications from "@rilldata/web-local/lib/components/notifications";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function renameFileArtifact(
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
  await dataModelerService.dispatch("renameEntity", [type, fromName, toName]);
  goto(getRouteFromName(toName, type), {
    replaceState: true,
  });
  notifications.send({
    message: `renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
  return invalidateAfterReconcile(queryClient, instanceId, resp);
}

export async function deleteFileArtifact(
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
    if (activeEntity.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }
    // Temporary until nodejs is removed
    await dataModelerService.dispatch("deleteEntity", [type, name]);

    return invalidateAfterReconcile(queryClient, instanceId, resp);
  } catch (err) {
    console.error(err);
  }
}
