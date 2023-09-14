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
import {
  createRuntimeServicePutFile,
  createRuntimeServiceDeleteFile,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function createFileSaver() {
  const saveFileMutation = createRuntimeServicePutFile();

  return async (path: string, blob: string) => {
    return get(saveFileMutation).mutateAsync({
      instanceId: get(runtime).instanceId,
      path,
      data: {
        blob,
        create: false,
        createOnly: false,
      },
    });
  };
}

export function createFileDeleter(
  entityNamesQuery: CreateQueryResult<Array<string>>
) {
  const deleteFileMutation = createRuntimeServiceDeleteFile();

  return async (
    name: string,
    type: EntityType,
    path = getFilePathFromNameAndType(name, type)
  ) => {
    await get(deleteFileMutation).mutateAsync({
      instanceId: get(runtime).instanceId,
      path,
    });

    notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });

    // only redirect if the deleted entity is in focus
    if (get(appScreen)?.name !== name) return;

    const route = getRouteFromName(
      getNextEntityName(get(entityNamesQuery).data, name),
      type
    );
    /** set the href store so the menu selection has an immediate visual update. */
    currentHref.set(route);
    goto(route);
  };
}
