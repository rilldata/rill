import { goto } from "$app/navigation";
import { notifications } from "@rilldata/web-common/components/notifications";
import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
import {
  runtimeServiceDeleteFile,
  runtimeServiceRenameFile,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { removeLeadingSlash } from "./entity-mappers";

export async function renameFileArtifact(
  instanceId: string,
  fromPath: string,
  toPath: string,
) {
  await runtimeServiceRenameFile(instanceId, {
    fromPath,
    toPath,
  });

  const fromName = extractFileName(fromPath);
  const toName = extractFileName(toPath);

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${fromName} to ${toName}`,
  });
}

export async function deleteFileArtifact(
  instanceId: string,
  filePath: string,
  showNotification = true,
) {
  const name = extractFileName(filePath);
  try {
    await runtimeServiceDeleteFile(instanceId, removeLeadingSlash(filePath));

    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${name}` });
    }

    await goto("/");
  } catch (err) {
    console.error(err);
  }
}
