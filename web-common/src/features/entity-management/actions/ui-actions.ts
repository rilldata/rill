import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions/actions.ts";
import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";
import {
  INVALID_NAME_MESSAGE,
  isDuplicateName,
  VALID_NAME_PATTERN,
} from "@rilldata/web-common/features/entity-management/name-utils.ts";
import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing.ts";
import { extractErrorMessage } from "@rilldata/web-common/lib/errors.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient.ts";
import { getFileNamesInDirectory } from "../file-selectors.ts";
import {
  matchReadonlyFile,
  type ReadonlyMatcher,
} from "@rilldata/web-common/features/entity-management/actions/readonly-files.ts";

export async function handleEntityRename(
  client: RuntimeClient,
  newName: string,
  existingPath: string,
  existingName: string,
  readonlyExtras: ReadonlyMatcher[] = [],
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
    client,
    folder,
  );
  if (isDuplicateName(newName, existingName, fileNamesInDirectory)) {
    eventBus.emit("notification", {
      message: `Name ${newName} is already in use`,
    });

    return;
  }

  const newFilePath = (folder ? `${folder}/` : "/") + newName;
  if (matchReadonlyFile(newFilePath, readonlyExtras)) {
    eventBus.emit("notification", {
      message: `Cannot rename to ${newFilePath}. It is a protected path.`,
    });
    return;
  }

  // Rename the file
  try {
    await renameFileArtifact(client, existingPath, newFilePath);

    return getFileHref(`/${removeLeadingSlash(newFilePath)}`);
  } catch (err) {
    console.error(extractErrorMessage(err));
  }
}
