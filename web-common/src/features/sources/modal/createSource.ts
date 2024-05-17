import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";

export async function createSource(
  instanceId: string,
  tableName: string,
  yaml: string,
) {
  return runtimeServicePutFile(instanceId, {
    path: getFileAPIPathFromNameAndType(tableName, EntityType.Table),
    blob: yaml,
    // create source is used to upload and replace.
    // so we cannot send createOnly=true until we refactor it to use refresh source
    createOnly: false,
  });
}
