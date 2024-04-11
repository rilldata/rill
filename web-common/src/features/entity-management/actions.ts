import { notifications } from "@rilldata/web-common/components/notifications";
import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
import {
  runtimeServiceDeleteFile,
  runtimeServiceRenameFile,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { getLabel, removeLeadingSlash } from "./entity-mappers";
import type { EntityType } from "./types";

export async function renameFileArtifact(
  instanceId: string,
  fromPath: string,
  toPath: string,
  type: EntityType,
) {
  await runtimeServiceRenameFile(instanceId, {
    fromPath,
    toPath,
  });

  const fromName = extractFileName(fromPath);
  const toName = extractFileName(toPath);

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });
}

export async function deleteFileArtifact(
  instanceId: string,
  filePath: string,
  type: EntityType,
  showNotification = true,
) {
  const name = extractFileName(filePath);
  try {
    await runtimeServiceDeleteFile(instanceId, removeLeadingSlash(filePath));

    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });
    }
  } catch (err) {
    console.error(err);
  }
}
