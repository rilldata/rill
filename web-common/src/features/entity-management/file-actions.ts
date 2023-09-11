import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import type { EntityType } from "@rilldata/web-common/features/entity-management/types";
import type { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export async function saveFile(
  instanceId: string,
  name: string,
  type: EntityType,
  blob: string,
  saveMutation: ReturnType<typeof createRuntimeServicePutFile>
) {
  const filePath = getFilePathFromNameAndType(name, type);

  await get(saveMutation).mutateAsync({
    instanceId,
    data: {
      blob,
      create: false,
      createOnly: false,
    },
    path: filePath,
  });
}
