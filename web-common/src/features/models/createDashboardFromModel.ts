import { goto } from "$app/navigation";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { generateDashboardYAMLForModelV2 } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import type { TelemetryParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import {
  createRuntimeServicePutFile,
  queryServiceTableColumns,
} from "@rilldata/web-common/runtime-client";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function createDashboardFromModelCreator(
  allNamesQuery: CreateQueryResult<Array<string>>,
  navigationTelemetryParams: TelemetryParams | undefined
) {
  const putFileMutation = createRuntimeServicePutFile();

  return async (model: V1Resource, sourceName: string, pathPrefix?: string) => {
    pathPrefix ??= "/dashboard/";

    const instanceId = get(runtime).instanceId;
    const columnsResp = await queryServiceTableColumns(
      instanceId,
      model.meta.name.name
    );

    const yaml = generateDashboardYAMLForModelV2(
      model.meta.name.name,
      columnsResp.profileColumns
    );

    const newDashboardName = getName(
      `${sourceName}_dashboard`,
      get(allNamesQuery).data ?? []
    );

    await get(putFileMutation).mutateAsync({
      instanceId,
      path: `${pathPrefix}${newDashboardName}`,
      data: {
        blob: yaml,
        create: true,
        createOnly: true,
      },
    });

    if (!navigationTelemetryParams) return;
    goto(`/dashboard/${newDashboardName}`);
    behaviourEvent.fireNavigationEvent(
      newDashboardName,
      navigationTelemetryParams.medium,
      navigationTelemetryParams.space,
      // TODO: existing code does this as well.
      //       should it be the screen when the original action was triggered?
      get(appScreen).type,
      MetricsEventScreenName.Dashboard
    );
  };
}
