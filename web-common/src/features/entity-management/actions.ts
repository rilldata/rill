import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  extractFileName,
  getTopLevelFolder,
  splitFolderFileNameAndExtension,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import { fileIsMainEntity } from "@rilldata/web-common/features/entity-management/file-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { isNotFoundError } from "@rilldata/web-common/lib/errors";
import {
  runtimeServiceDeleteFile,
  runtimeServiceGetFile,
  runtimeServiceGetResource,
  runtimeServicePutFile,
  runtimeServiceRenameFile,
  type RuntimeServicePutFileBody,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
  client: RuntimeClient,
  runtimeServicePutFileBody: RuntimeServicePutFileBody,
) {
  const projectParserStartingVersion = getProjectParserVersion(
    client.instanceId,
  );

  await runtimeServicePutFile(client, runtimeServicePutFileBody);

  // Wait for the file to be processed by the parser
  await waitForProjectParserVersion(
    client.instanceId,
    projectParserStartingVersion + 1,
  );
}

// Resource-level reconciliation
export async function waitForResourceReconciliation(
  client: RuntimeClient,
  resourceName: string,
  resourceKind: ResourceKind,
) {
  const pollInterval = 2_000; // 2 seconds
  let attempt = 0;

  while (true) {
    attempt++;
    try {
      const resource = await runtimeServiceGetResource(client, {
        name: { kind: resourceKind, name: resourceName },
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
      if (isNotFoundError(error)) {
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

/**
 * Polls for file creation until the file exists or the operation is cancelled.
 * Returns true if file was found, false if cancelled.
 */
export async function pollForFileCreation(
  client: RuntimeClient,
  filePath: string,
  abortSignal: AbortSignal,
  pollIntervalMs: number = 1000,
): Promise<boolean> {
  while (!abortSignal.aborted) {
    await new Promise((resolve) => setTimeout(resolve, pollIntervalMs));

    try {
      await runtimeServiceGetFile(client, { path: filePath });
      return true; // success, file exists
    } catch {
      // 404 error, file is not ready yet, continue polling
    }
  }
  return false; // cancelled
}

export async function renameFileArtifact(
  client: RuntimeClient,
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
    await runtimeServiceRenameFile(client, {
      fromPath,
      toPath,
    });

    client.requestQueue.removeByName(fromName);
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
      message: `Failed to rename ${fromName} to ${toName}: ${extractMessage(err.rawMessage ?? err.response?.data?.message ?? err.message)}`,
    });
  }
}

export async function duplicateFileArtifact(
  client: RuntimeClient,
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
  await runtimeServicePutFile(client, {
    path: newFilePath,
    blob: fileContent ?? "",
  });

  // Return the new file path
  return newFilePath;
}

export async function deleteFileArtifact(
  client: RuntimeClient,
  filePath: string,
  force = false,
) {
  const name = extractFileName(filePath);
  try {
    await runtimeServiceDeleteFile(client, {
      path: filePath,
      force,
    });

    client.requestQueue.removeByName(name);
  } catch (err) {
    eventBus.emit("notification", {
      message: `Failed to delete ${name}: ${extractMessage(err.rawMessage ?? err.response?.data?.message ?? err.message)}`,
    });
  }
}

function extractMessage(msg: string) {
  if (msg.endsWith("directory not empty")) return "directory not empty";
  return msg;
}
