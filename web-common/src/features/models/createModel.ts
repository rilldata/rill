import { goto } from "$app/navigation";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import type { TelemetryParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export function createModelCreator(
  navigationTelemetryParams?: TelemetryParams
) {
  const createFileMutation = createRuntimeServicePutFile();

  // getting the pathPrefix from the argument makes it easy to add folders
  return async (newModelName: string, pathPrefix: string, sql: string) => {
    await get(createFileMutation).mutateAsync({
      instanceId: get(runtime).instanceId,
      path: `${pathPrefix}${newModelName}.sql`,
      data: {
        blob: sql,
        createOnly: true,
      },
    });
    if (!navigationTelemetryParams) return;

    goto(`/model/${newModelName}?focus`);
    behaviourEvent.fireNavigationEvent(
      newModelName,
      navigationTelemetryParams.medium,
      navigationTelemetryParams.space,
      get(appScreen).type,
      MetricsEventScreenName.Model
    );
  };
}
