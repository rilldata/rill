import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { currentHref } from "@rilldata/web-common/layout/navigation/stores";
import {
  runtimeServiceDeleteFile,
  runtimeServiceRenameFile,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { get } from "svelte/store";
import {
  getFileAPIPathFromNameAndType,
  getLabel,
  getRouteFromName,
  removeLeadingSlash,
} from "./entity-mappers";
import { getNextEntityName } from "./name-utils";
import type { EntityType } from "./types";

export async function renameFileArtifact(
  instanceId: string,
  fromName: string,
  toName: string,
  type: EntityType,
) {
  await runtimeServiceRenameFile(instanceId, {
    fromPath: getFileAPIPathFromNameAndType(fromName, type),
    toPath: getFileAPIPathFromNameAndType(toName, type),
  });

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
}

export async function deleteFileArtifact(
  instanceId: string,
  name: string,
  type: EntityType,
  names: Array<string>,
  showNotification = true,
) {
  const path = getFileAPIPathFromNameAndType(name, type);
  try {
    await runtimeServiceDeleteFile(instanceId, removeLeadingSlash(path));

    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });
    }

    if (get(appScreen)?.name === name) {
      const route = getRouteFromName(getNextEntityName(names, name), type);
      /** set the href store so the menu selection has an immediate visual update. */
      currentHref.set(route);
      goto(route);
    }
  } catch (err) {
    console.error(err);
  }
}
