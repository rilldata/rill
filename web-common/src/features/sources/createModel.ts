import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { createModel } from "@rilldata/web-common/features/models/createModel";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { notifications } from "../../components/notifications";
import { runtimeServicePutFile } from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";
import { getFileAPIPathFromNameAndType } from "../entity-management/entity-mappers";
import { EntityType } from "../entity-management/types";
import { getModelNames } from "../models/selectors";

// TODO: merge these 2 methods
export async function createModelFromSource(
  instanceId: string,
  modelNames: Array<string>,
  sourceName: string,
  sourceNameInQuery: string,
): Promise<string> {
  const newModelName = getName(`${sourceName}_model`, modelNames);
  await createModel(
    instanceId,
    newModelName,
    `select * from ${sourceNameInQuery}`,
  );
  notifications.send({
    message: `Queried ${sourceNameInQuery} in workspace`,
  });
  return newModelName;
}

export async function createModelFromSourceV2(
  queryClient: QueryClient,
  sourceName: string,
  tableName: string,
  folder: string,
  notify = false,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get new model name
  const modelNames = await getModelNames(queryClient, instanceId);
  const newModelName = getName(`${sourceName}_model`, modelNames);
  const newModelPath = `${folder}/${newModelName}.sql`;

  // Create model
  await runtimeServicePutFile(instanceId, newModelPath, {
    blob: `-- @kind: model
select * from ${tableName}`,
    createOnly: true,
  });

  if (notify) {
    notifications.send({
      message: `Queried ${tableName} in workspace`,
    });
  }

  // Done
  return [newModelPath, newModelName];
}
