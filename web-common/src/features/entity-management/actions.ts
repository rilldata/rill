import { notifications } from "@rilldata/web-common/components/notifications";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { fileIsMainEntity } from "@rilldata/web-common/features/entity-management/file-selectors";
import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
import {
  runtimeServiceDeleteFile,
  runtimeServiceRenameFile,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { addLeadingSlash, removeLeadingSlash } from "./entity-mappers";
import { get } from "svelte/store";

export async function renameFileArtifact(
  instanceId: string,
  fromPath: string,
  toPath: string,
) {
  const fromName = extractFileName(fromPath);
  const toName = extractFileName(toPath);

  if (fileIsMainEntity(fromPath)) {
    // try and copy over kind+name proactively for main entities (.sql,.yml,.yaml)
    // this fixes a flicker when renamed
    const fromFileArtifact = fileArtifacts.getFileArtifact(
      addLeadingSlash(fromPath),
    );
    const toFileArtifact = fileArtifacts.getFileArtifact(
      addLeadingSlash(toPath),
    );
    if (!get(toFileArtifact.name)) {
      // if there is no name set yet copy over from the source
      toFileArtifact.name.set(get(fromFileArtifact.name));
    }
  }

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
