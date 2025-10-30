import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  extractFileName,
  getTopLevelFolder,
  splitFolderFileNameAndExtension,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import { fileIsMainEntity } from "@rilldata/web-common/features/entity-management/file-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import {
  runtimeServiceDeleteFile,
  runtimeServiceGetResource,
  runtimeServicePutFile,
  runtimeServiceRenameFile,
  type RuntimeServicePutFileBody,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { get } from "svelte/store";
import {
  FolderNameToResourceKind,
  addLeadingSlash,
  removeLeadingSlash,
} from "./entity-mappers";
import {
  getProjectParserVersion,
  waitForProjectParserVersion,
} from "./project-parser";
import { ResourceKind } from "./resource-selectors";

export async function runtimeServicePutFileAndWaitForReconciliation(
  instanceId: string,
  runtimeServicePutFileBody: RuntimeServicePutFileBody,
) {
  const projectParserStartingVersion = getProjectParserVersion(instanceId);

  await runtimeServicePutFile(instanceId, runtimeServicePutFileBody);

  // Wait for the file to be processed by the parser
  await waitForProjectParserVersion(
    instanceId,
    projectParserStartingVersion + 1,
  );
}

// Resource-level reconciliation
export async function waitForResourceReconciliation(
  instanceId: string,
  resourceName: string,
  resourceKind: ResourceKind,
) {
  const pollInterval = 2_000; // 2 seconds
  let attempt = 0;

  while (true) {
    attempt++;
    try {
      const resource = await runtimeServiceGetResource(instanceId, {
        "name.kind": resourceKind,
        "name.name": resourceName,
      });

      // Check if there's a reconcile error
      if (resource.resource?.meta?.reconcileError) {
        const error = new Error("Resource configuration failed to reconcile");
        (error as any).details = resource.resource.meta.reconcileError;
        throw error;
      }

      // Check the reconcile status
      const reconcileStatus = resource.resource?.meta?.reconcileStatus;
      if (reconcileStatus === "RECONCILE_STATUS_IDLE") {
        return; // Success!
      }

      // Still reconciling, continue polling
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
      continue;
    } catch (error) {
      // Resource not found could mean it was deleted due to reconcile failure
      if (error?.status === 404 || error?.response?.status === 404) {
        if (attempt >= 3) {
          // After 6 seconds, assume reconcile failure
          throw new Error(
            `Resource configuration failed to reconcile and was automatically deleted. This usually indicates a connection or configuration error.`,
          );
        }

        // Wait and try again
        await new Promise((resolve) => setTimeout(resolve, pollInterval));
        continue;
      }

      // Re-throw other errors
      throw error;
    }
  }
}

export async function renameFileArtifact(
  instanceId: string,
  fromPath: string,
  toPath: string,
) {
  const fromName = extractFileName(fromPath);
  const toName = extractFileName(toPath);

  const fromFileArtifact = fileArtifacts.getFileArtifact(
    addLeadingSlash(fromPath),
  );
  const fromResName = get(fromFileArtifact.resourceName);

  if (fileIsMainEntity(fromPath)) {
    // try and copy over kind+name proactively for main entities (.sql,.yml,.yaml)
    // this fixes a flicker when renamed
    const toFileArtifact = fileArtifacts.getFileArtifact(
      addLeadingSlash(toPath),
    );
    if (!get(toFileArtifact.resourceName)) {
      // if there is no name set yet copy over from the source
      toFileArtifact.resourceName.set(fromResName);
    }
  }

  try {
    await runtimeServiceRenameFile(instanceId, {
      fromPath,
      toPath,
    });

    httpRequestQueue.removeByName(fromName);
    const topLevelFromFolder = getTopLevelFolder(addLeadingSlash(fromPath));
    const topLevelToFolder = getTopLevelFolder(addLeadingSlash(toPath));

    if (
      fromResName?.kind &&
      topLevelFromFolder !== topLevelToFolder &&
      FolderNameToResourceKind[removeLeadingSlash(topLevelFromFolder)] ===
        fromResName?.kind &&
      !toPath.endsWith(".sql")
    ) {
      eventBus.emit("notification", {
        message: `Moving ${fromName} out of its native folder. Make sure to specify the resource type with the "type" key.`,
      });
    }
  } catch (err) {
    eventBus.emit("notification", {
      message: `Failed to rename ${fromName} to ${toName}: ${extractMessage(err.response?.data?.message ?? err.message)}`,
    });
  }
}

export async function duplicateFileArtifact(
  instanceId: string,
  filePath: string,
): Promise<string> {
  // Get new file path
  const [folder, fileName, extension] =
    splitFolderFileNameAndExtension(filePath);
  const newFilePath = `${folder}/${fileName} (copy)${extension}`;

  // Get file content
  const fileArtifact = fileArtifacts.getFileArtifact(filePath);
  await fileArtifact.fetchContent();
  const fileContent = get(fileArtifact.remoteContent);

  // Create new file
  await runtimeServicePutFile(instanceId, {
    path: newFilePath,
    blob: fileContent ?? "",
  });

  // Return the new file path
  return newFilePath;
}

export async function deleteFileArtifact(
  instanceId: string,
  filePath: string,
  force = false,
) {
  const name = extractFileName(filePath);
  try {
    await runtimeServiceDeleteFile(instanceId, {
      path: filePath,
      force,
    });

    httpRequestQueue.removeByName(name);
  } catch (err) {
    eventBus.emit("notification", {
      message: `Failed to delete ${name}: ${extractMessage(err.response?.data?.message ?? err.message)}`,
    });
  }
}

function extractMessage(msg: string) {
  if (msg.endsWith("directory not empty")) return "directory not empty";
  return msg;
}
