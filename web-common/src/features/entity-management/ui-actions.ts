import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { queryClient } from "../../lib/svelte-query/globalQueryClient";
import { getFileNamesInDirectory } from "./file-selectors";

export async function handleEntityRename(
  instanceId: string,
  newName: string,
  existingPath: string,
  existingName: string,
) {
  const [folder] = splitFolderAndFileName(existingPath);

  // Check if the new name is valid
  if (!newName.match(VALID_NAME_PATTERN)) {
    eventBus.emit("notification", {
      message: INVALID_NAME_MESSAGE,
    });

    return;
  }

  // Check if the new name is already in use
  const fileNamesInDirectory = await getFileNamesInDirectory(
    queryClient,
    instanceId,
    folder,
  );
  if (isDuplicateName(newName, existingName, fileNamesInDirectory)) {
    eventBus.emit("notification", {
      message: `Name ${newName} is already in use`,
    });

    return;
  }

  // Rename the file
  try {
    const newFilePath = (folder ? `${folder}/` : "/") + newName;

    await renameFileArtifact(instanceId, existingPath, newFilePath);

    return `/files/${removeLeadingSlash(newFilePath)}`;
  } catch (err) {
    console.error(err.response?.data?.message ?? err);
  }
}
