import { goto } from "$app/navigation";
import {
  EntityAction,
  entityActionQueueStore,
} from "@rilldata/web-common/features/entity-management/entity-action-queue";
import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export async function createSource(
  instanceId: string,
  tableName: string,
  yaml: string,
  behaviourEventMedium = BehaviourEventMedium.Button
) {
  entityActionQueueStore.add(tableName, {
    action: EntityAction.Create,
    screenName: get(appScreen),
    behaviourEventMedium,
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
