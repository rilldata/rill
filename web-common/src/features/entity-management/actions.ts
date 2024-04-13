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
  const fromName = extractFileName(fromPath);
  const toName = extractFileName(toPath);

  try {
    await runtimeServiceRenameFile(instanceId, {
      fromPath,
      toPath,
    });

    httpRequestQueue.removeByName(fromName);
  } catch (err) {
    notifications.send({
      message: `Failed to rename ${fromName} to ${toName}: ${extractMessage(err.response?.data?.message ?? err.message)}`,
    });
  }
}

export async function deleteFileArtifact(instanceId: string, filePath: string) {
  const name = extractFileName(filePath);
  try {
    await runtimeServiceDeleteFile(instanceId, removeLeadingSlash(filePath));

    httpRequestQueue.removeByName(name);
  } catch (err) {
    notifications.send({
      message: `Failed to delete ${name}: ${extractMessage(err.response?.data?.message ?? err.message)}`,
    });
  }
}

function extractMessage(msg: string) {
  if (msg.endsWith("directory not empty")) return "directory not empty";
  return msg;
}
