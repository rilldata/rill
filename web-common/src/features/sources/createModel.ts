import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { get } from "svelte/store";
import { notifications } from "../../components/notifications";
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

  // Create model
  await runtimeServicePutFile(instanceId, newModelPath, {
    blob: `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

-- @kind: model

select * from ${tableName}`,
    createOnly: true,
  });

  if (notify) {
    notifications.send({
      message: `Queried ${tableName} in workspace`,
    });
  }

  // Done
  return ["/" + newModelPath, newModelName];
}
