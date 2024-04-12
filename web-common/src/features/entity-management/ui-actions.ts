import { notifications } from "@rilldata/web-common/components/notifications";
import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { fetchAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export async function handleEntityRename(
  instanceId: string,
  target: HTMLInputElement,
  existingPath: string,
  existingName: string,
) {
  const [folder, fileName] = splitFolderAndName(existingPath);
  const suffix = fileName.replace(existingName, "");

  if (!target.value.match(VALID_NAME_PATTERN)) {
    notifications.send({
      message: INVALID_NAME_MESSAGE,
    });
    target.value = existingName; // resets the input
    return;
  }

  const allNames = await fetchAllNames(queryClient, instanceId);

  if (isDuplicateName(target.value, existingName, allNames)) {
    notifications.send({
      message: `Name ${target.value} is already in use`,
    });
    target.value = existingName; // resets the input
    return;
  }

  try {
    const toName = target.value;

    const newFilePath = (folder ? `${folder}/` : "") + toName + suffix;

    await renameFileArtifact(instanceId, existingPath, newFilePath);

    return `/files/${newFilePath}`;
  } catch (err) {
    console.error(err.response?.data?.message ?? err);
  }
}
