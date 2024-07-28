import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { get } from "svelte/store";
import { hasSpaces } from "../../lib/string-utils";
import { runtimeServicePutFile } from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";

export async function createModelFromSource(
  sourceName: string,
  tableName: string,
  folder: string,
  notify = false,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  folder = removeLeadingSlash(folder);

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${sourceName}_model`, allNames);
  const newModelPath = `${folder}/${newModelName}.sql`;

  // Compile model SQL
  const topOfFile = `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models`;
  const selectStatement = hasSpaces(tableName)
    ? `select * from "${tableName}"`
    : `select * from ${tableName}`;
  const modelSQL = `${topOfFile}\n\n${selectStatement}`;

  // Create model
  await runtimeServicePutFile(instanceId, {
    path: newModelPath,
    blob: modelSQL,
    createOnly: true,
  });

  if (notify) {
    eventBus.emit("notification", {
      message: `Queried ${tableName} in workspace`,
    });
  }

  // Done
  return ["/" + newModelPath, newModelName];
}
