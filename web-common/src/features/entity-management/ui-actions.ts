import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import {
  extractFileName,
  splitFolderAndFileName,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

export async function handleEntityRename(
  instanceId: string,
  newName: string,
  existingPath: string,
  existingName: string,
  allNames: string[],
) {
  const [folder] = splitFolderAndFileName(existingPath);

  if (!newName.match(VALID_NAME_PATTERN)) {
    eventBus.emit("notification", {
      message: INVALID_NAME_MESSAGE,
    });

    return;
  }

  if (isDuplicateName(extractFileName(newName), existingName, allNames)) {
    eventBus.emit("notification", {
      message: `Name ${newName} is already in use`,
    });

    return;
  }

  try {
    const newFilePath = (folder ? `${folder}/` : "/") + newName;

    await renameFileArtifact(instanceId, existingPath, newFilePath);

    return `/files/${removeLeadingSlash(newFilePath)}`;
  } catch (err) {
    console.error(err.response?.data?.message ?? err);
  }
}
