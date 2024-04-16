import { goto } from "$app/navigation";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { getFileAPIPathFromNameAndType } from "../entity-management/entity-mappers";

export async function createModel(
  instanceId: string,
  newModelName: string,
  sql = "",
) {
  await runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(newModelName, EntityType.Model),
    {
      blob: sql,
      createOnly: true,
    },
  );
  await goto(`/files/models/${newModelName}?focus`);
}
