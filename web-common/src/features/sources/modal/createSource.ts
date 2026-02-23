import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { sourceIngestionTracker } from "../sources-store";

export async function createSource(
  instanceId: string,
  tableName: string,
  yaml: string,
) {
  const filePath = getFileAPIPathFromNameAndType(tableName, EntityType.Table);
  const normalizedPath = `/${filePath}`;
  sourceIngestionTracker.trackPending(normalizedPath);
  try {
    return await runtimeServicePutFile(instanceId, {
      path: filePath,
      blob: yaml,
      // create source is used to upload and replace.
      // so we cannot send createOnly=true until we refactor it to use refresh source
      createOnly: false,
    });
  } catch (error) {
    sourceIngestionTracker.trackCancelled(normalizedPath);
    throw error;
  }
}
