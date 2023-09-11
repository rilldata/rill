import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import {
  getFilePathFromNameAndType,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { isDuplicateName } from "@rilldata/web-common/features/entity-management/name-utils";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import type { createRuntimeServiceRenameFile } from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { get } from "svelte/types/runtime/store";

export async function validateAndRenameEntity(
  instanceId: string,
  fromName: string,
  toName: string,
  allNames: Array<string>,
  entityType: EntityType,
  renameMutation: ReturnType<typeof createRuntimeServiceRenameFile>
): Promise<boolean> {
  if (!toName.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
    notifications.send({
      message: `${getLabel(
        entityType
      )} name must start with a letter or underscore and contain only letters, numbers, and underscores`,
    });
    return false;
  }

  if (isDuplicateName(toName, fromName, allNames)) {
    notifications.send({
      message: `Name ${toName} is already in use`,
    });
    return false;
  }

  try {
    await renameFile(instanceId, fromName, toName, entityType, renameMutation);
  } catch (err) {
    console.error(err.response.data.message);
  }

  return true;
}

export async function renameFile(
  instanceId: string,
  fromName: string,
  toName: string,
  entityType: EntityType,
  renameMutation: ReturnType<typeof createRuntimeServiceRenameFile>
) {
  await get(renameMutation).mutateAsync({
    instanceId,
    data: {
      fromPath: getFilePathFromNameAndType(fromName, entityType),
      toPath: getFilePathFromNameAndType(toName, entityType),
    },
  });

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(entityType)} ${fromName} to ${toName}`,
  });

  const route = getRouteFromName(toName, entityType);
  goto(entityType === EntityType.MetricsDefinition ? `${route}/edit` : route, {
    replaceState: true,
  });
  // TODO: no telemetry for rename?
}
