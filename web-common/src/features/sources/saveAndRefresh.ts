import { get } from "svelte/store";
import { runtimeServicePutFile } from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";

export async function saveAndRefresh(filePath: string, yaml: string) {
  const instanceId = get(runtime).instanceId;

  return runtimeServicePutFile(instanceId, filePath, {
    blob: yaml,
    create: false,
    createOnly: false,
  });
}
