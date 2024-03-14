import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { getFileAPIPathFromNameAndType } from "../entity-management/entity-mappers";

export async function createCustomDashboard(
  instanceId: string,
  newCustomDashboardName: string,
) {
  await runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(newCustomDashboardName, EntityType.Dashboard),
    {
      blob: "abc",
      createOnly: true,
    },
  );
}
