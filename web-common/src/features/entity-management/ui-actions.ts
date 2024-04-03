import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { fetchAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { extractFileExtension } from "@rilldata/web-common/features/sources/extract-file-name";
import type { QueryClient } from "@tanstack/query-core";

export async function handleEntityRename(
  queryClient: QueryClient,
  instanceId: string,
  target: HTMLInputElement,
  filePath: string,
  entityType: EntityType, // temporary param
) {
  const [folder, fileName] = splitFolderAndName(filePath);
  const extension = extractFileExtension(filePath);

  if (!target.value.match(VALID_NAME_PATTERN)) {
    notifications.send({
      message: INVALID_NAME_MESSAGE,
    });
    target.value = fileName; // resets the input
    return;
  }
  const allNames = await fetchAllNames(queryClient, instanceId);
  if (isDuplicateName(target.value, fileName, allNames)) {
    notifications.send({
      message: `Name ${target.value} is already in use`,
    });
    target.value = fileName; // resets the input
    return;
  }

  try {
    const toName = target.value;
    await renameFileArtifact(
      instanceId,
      filePath,
      (folder ? `${folder}/` : "") + toName + extension,
      entityType,
    );
    // TODO: replace once we have asset explorer
    void goto(getRouteFromName(toName, entityType), {
      replaceState: true,
    });
  } catch (err) {
    console.error(err.response.data.message);
  }
}
