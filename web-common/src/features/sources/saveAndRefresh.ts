import { get } from "svelte/store";
import { runtimeServicePutFile } from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";
import { getFileAPIPathFromNameAndType } from "../entity-management/entity-mappers";
import { EntityType } from "../entity-management/types";

export async function saveAndRefresh(tableName: string, yaml: string) {
  const instanceId = get(runtime).instanceId;
  const filePath = getFileAPIPathFromNameAndType(tableName, EntityType.Table);

  return runtimeServicePutFile(instanceId, filePath, {
    blob: yaml,
    create: false,
    createOnly: false,
  });
}
