import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

export async function handleEntityRename(
  instanceId: string,
  target: HTMLInputElement,
  existingPath: string,
  existingName: string,
  allNames: string[],
) {
  const [folder] = splitFolderAndName(existingPath);

  if (!target.value.match(VALID_NAME_PATTERN)) {
    eventBus.emit("notification", {
      message: INVALID_NAME_MESSAGE,
    });
    target.value = existingName; // resets the input
    return;
  }

  if (isDuplicateName(target.value, existingName, allNames)) {
    eventBus.emit("notification", {
      message: `Name ${target.value} is already in use`,
    });
    target.value = existingName; // resets the input
    return;
  }

  try {
    const toName = target.value;

    const newFilePath = (folder ? `${folder}/` : "/") + toName;

    await renameFileArtifact(instanceId, existingPath, newFilePath);

    return `/files/${removeLeadingSlash(newFilePath)}`;
  } catch (err) {
    console.error(err.response?.data?.message ?? err);
  }
}
