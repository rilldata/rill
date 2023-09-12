import { goto } from "$app/navigation";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";

export async function createModel(newModelName: string, sql = "") {
  await runtimeServicePutFile(
    get(runtime).instanceId,
    getFilePathFromNameAndType(newModelName, EntityType.Model),
    {
      blob: sql,
      createOnly: true,
    }
  );
  goto(`/model/${newModelName}?focus`);
}
