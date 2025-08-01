import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  getName,
  isNonStandardIdentifier,
} from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  runtimeServiceAnalyzeConnectors,
  runtimeServicePutFile,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { makeSufficientlyQualifiedTableName } from "../olap/olap-config";

export async function createYamlModelFromTable(
  queryClient: QueryClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get driver name
  const analyzeConnectorsQueryKey =
    getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId);
  const analyzeConnectorsQueryFn = async () =>
    runtimeServiceAnalyzeConnectors(instanceId);
  const connectors = await queryClient.fetchQuery({
    queryKey: analyzeConnectorsQueryKey,
    queryFn: analyzeConnectorsQueryFn,
  });
  const analyzedConnector = connectors?.connectors?.find(
    (c) => c.name === connector,
  );
  if (!analyzedConnector) {
    throw new Error(`Could not find connector ${connector}`);
  }
  const driverName = analyzedConnector.driver?.name as string;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.yaml`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    driverName,
    database,
    databaseSchema,
    table,
  );

  // Create YAML model content
  const selectStatement = isNonStandardIdentifier(
    sufficientlyQualifiedTableName,
  )
    ? `select * from "${sufficientlyQualifiedTableName}"`
    : `select * from ${sufficientlyQualifiedTableName}`;

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
  eventBus.fire("file-created", {
    path: newModelPath,
    kind: ResourceKind.Model,
  });

  return [newModelPath, newModelName];
}
