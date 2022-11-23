import { goto } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  V1DeleteFileAndMigrateResponse,
  V1RenameFileAndMigrateResponse,
} from "@rilldata/web-common/runtime-client";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getNextEntityName } from "@rilldata/web-local/common/utils/getNextEntityId";
import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import type { RuntimeState } from "@rilldata/web-local/lib/application-state-stores/application-store";
import {
  getFileFromName,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-local/lib/components/entity-mappers/mappers";
import notifications from "@rilldata/web-local/lib/components/notifications";
import { queryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
import type { UseMutationResult } from "@sveltestack/svelte-query";

export async function renameEntity(
  runtimeState: RuntimeState,
  fromName: string,
  toName: string,
  type: EntityType,
  renameMutation: UseMutationResult<V1RenameFileAndMigrateResponse>
) {
  await renameMutation.mutateAsync({
    data: {
      repoId: runtimeState.repoId,
      instanceId: runtimeState.instanceId,
      fromPath: getFileFromName(fromName, type),
      toPath: getFileFromName(toName, type),
    },
  });
  await dataModelerService.dispatch("renameEntity", [type, fromName, toName]);
  goto(getRouteFromName(toName, type), {
    replaceState: true,
  });
  notifications.send({
    message: `renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
  await queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(runtimeState.repoId)
  );
}

export async function deleteEntity(
  runtimeState: RuntimeState,
  name: string,
  type: EntityType,
  deleteMutation: UseMutationResult<V1DeleteFileAndMigrateResponse>,
  activeEntity: ActiveEntity,
  names: Array<string>
) {
  try {
    await deleteMutation.mutateAsync({
      data: {
        repoId: runtimeState.repoId,
        instanceId: runtimeState.instanceId,
        path: getFileFromName(name, type),
      },
    });
    if (activeEntity.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }
    // Temporary until nodejs is removed
    await dataModelerService.dispatch("deleteEntity", [type, name]);

    // TODO: update all entities based on affected path
    return queryClient.invalidateQueries(
      getRuntimeServiceListFilesQueryKey(runtimeState.repoId)
    );
  } catch (err) {
    console.error(err);
  }
}
