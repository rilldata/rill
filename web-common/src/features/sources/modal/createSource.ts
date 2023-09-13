import { goto } from "$app/navigation";
import {
  EntityAction,
  entityActionQueueStore,
} from "@rilldata/web-common/features/entity-management/entity-action-queue";
import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
import {
  createRuntimeServicePutFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export function createSourceCreator(
  behaviourEventMedium: BehaviourEventMedium
) {
  const putFileMutation = createRuntimeServicePutFile();

  return async (tableName: string, yaml: string, pathPrefix?: string) => {
    pathPrefix ??= "/sources";

    entityActionQueueStore.add(tableName, EntityAction.Create, {
      space: MetricsEventSpace.Modal,
      screenName: get(appScreen).type,
      medium: behaviourEventMedium,
    });

    await get(putFileMutation).mutateAsync({
      instanceId: get(runtime).instanceId,
      path: `${pathPrefix}${tableName}.yaml`,
      data: {
        blob: yaml,
        create: true,
        createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
      },
    });

    // Navigate to source page
    goto(`/source/${tableName}`);
  };
}

export async function createSource(
  instanceId: string,
  tableName: string,
  yaml: string,
  behaviourEventMedium = BehaviourEventMedium.Button
) {
  entityActionQueueStore.add(tableName, EntityAction.Create, {
    space: MetricsEventSpace.Modal,
    screenName: get(appScreen).type,
    medium: behaviourEventMedium,
  });
  await runtimeServicePutFile(
    instanceId,
    getFilePathFromNameAndType(tableName, EntityType.Table),
    {
      blob: yaml,
      create: true,
      createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
    }
  );

  // Navigate to source page
  goto(`/source/${tableName}`);
}
