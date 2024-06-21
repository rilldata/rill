import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { get } from "svelte/store";
import { runtimeServicePutFile } from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { makeSufficientlyQualifiedTableName } from "./olap-config";

export async function createModelFromTable(
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.sql`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    connector,
    database,
    databaseSchema,
    table,
  );

  // Create model
  await runtimeServicePutFile(instanceId, {
    path: newModelPath,
    blob: `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

select * from ${sufficientlyQualifiedTableName}
{{ if dev }} limit 100000 {{ end}}`,
    createOnly: true,
  });

  eventBus.emit("notification", {
    message: `Queried ${table} in workspace`,
  });

  // Done
  return ["/" + newModelPath, newModelName];
}
