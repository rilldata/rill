import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  getName,
  isNonStandardIdentifier,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { runtimeServicePutFile } from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";

export async function createYamlModelFromTable(
  queryClient: QueryClient,
  connector: string,
  table: string,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.yaml`;

  // For YAML models, use just the table name since connector context is specified
  // The connector will handle the proper qualification based on its configuration
  const selectStatement = isNonStandardIdentifier(table)
    ? `select * from "${table}"`
    : `select * from ${table}`;

  const yamlContent = `connector: ${connector}
sql: ${selectStatement}`;

  // Write the YAML file
  await runtimeServicePutFile(instanceId, {
    path: newModelPath,
    blob: yamlContent,
  });

  // Invalidate relevant queries
  await queryClient.invalidateQueries({
    queryKey: ["runtimeServiceListFiles", instanceId],
  });

  // Fire event for file creation
  eventBus.emit("notification", {
    message: `Created YAML model from ${table}`,
  });

  return ["/" + newModelPath, newModelName];
}
