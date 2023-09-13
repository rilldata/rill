import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { createModelCreator } from "@rilldata/web-common/features/models/createModel";
import type { TelemetryParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { notifications } from "../../components/notifications";

export function createModelFromSourceCreator(
  allNamesQuery: CreateQueryResult<Array<string>>,
  telemetryParams?: TelemetryParams
) {
  const modelCreator = createModelCreator(telemetryParams);

  // getting the pathPrefix from the argument makes it easy to add folders
  return async (
    source: V1Resource,
    sourceName: string,
    pathPrefix?: string
  ) => {
    pathPrefix ??= "/models/";

    const newModelName = getName(
      `${sourceName}_model`,
      get(allNamesQuery).data
    );
    await modelCreator(newModelName, pathPrefix, `select * from ${sourceName}`);
    notifications.send({
      message: `Queried ${sourceName} in workspace`,
    });
  };
}
