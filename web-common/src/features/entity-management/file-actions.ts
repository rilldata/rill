import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import {
  getFilePathFromNameAndType,
  getLabel,
  getRouteFromName,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { getNextEntityName } from "@rilldata/web-common/features/entity-management/name-utils";
import type { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { currentHref } from "@rilldata/web-common/layout/navigation/stores";
import { runtimeServiceDeleteFile } from "@rilldata/web-common/runtime-client";
import type { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export async function saveFile(
  instanceId: string,
  name: string,
  type: EntityType,
  blob: string,
  saveMutation: ReturnType<typeof createRuntimeServicePutFile>
) {
  const filePath = getFilePathFromNameAndType(name, type);

  await get(saveMutation).mutateAsync({
    instanceId,
    data: {
      blob,
      create: false,
      createOnly: false,
    },
    path: filePath,
  });
}

export async function deleteFile(
  instanceId: string,
  name: string,
  type: EntityType,
  names: Array<string>
) {
  const path = getFilePathFromNameAndType(name, type);
  await runtimeServiceDeleteFile(instanceId, path);

  notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });

  // only redirect if the deleted entity is in focus
  if (get(appScreen)?.name !== name) return;

  const route = getRouteFromName(getNextEntityName(names, name), type);
  /** set the href store so the menu selection has an immediate visual update. */
  currentHref.set(route);
  goto(route);
}
