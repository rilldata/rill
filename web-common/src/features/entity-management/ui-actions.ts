import { notifications } from "@rilldata/web-common/components/notifications";
import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { fetchAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { extractFileExtension } from "../sources/extract-file-name";

export async function handleEntityRename(
  instanceId: string,
  target: HTMLInputElement,
  existingPath: string,
  entityType: EntityType, // temporary param
) {
  const [folder, fileName] = splitFolderAndName(existingPath);
  const extension = extractFileExtension(existingPath);

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

    const newAPIPath = (folder ? `${folder}/` : "") + toName + extension;

    await renameFileArtifact(instanceId, existingPath, newAPIPath, entityType);

    return `/files/${newAPIPath}`;
  } catch (err) {
    console.error(err.response?.data?.message ?? err);
  }
}
