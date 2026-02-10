import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";

export async function createSource(
  instanceId: string,
  tableName: string,
  yaml: string,
) {
  return runtimeServicePutFile(instanceId, {
    path: `models/${tableName}.yaml`,
    blob: yaml,
    // create source is used to upload and replace.
    // so we cannot send createOnly=true until we refactor it to use refresh source
    createOnly: false,
  });
}
