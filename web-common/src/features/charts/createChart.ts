import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { getFileAPIPathFromNameAndType } from "../entity-management/entity-mappers";

export async function createChart(instanceId: string, newChartName: string) {
  await runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(newChartName, EntityType.Chart),
    {
      blob: "abc",
      createOnly: true,
    },
  );
}
