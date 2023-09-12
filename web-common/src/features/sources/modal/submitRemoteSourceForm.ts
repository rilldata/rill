import { goto } from "$app/navigation";
import { createSource } from "@rilldata/web-common/features/sources/modal/createSource";
import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import { appScreen } from "@rilldata/web-common/layout/app-store";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
import { runtimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants";
import { isProjectInitializedV2 } from "@rilldata/web-common/features/welcome/is-project-initialized";
import { compileCreateSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils";
import { fromYupFriendlyKey } from "@rilldata/web-common/features/sources/modal/yupSchemas";

export interface RemoteSourceFormValues {
  sourceName: string;
  [key: string]: any;
}

export async function submitRemoteSourceForm(
  queryClient: QueryClient,
  connectorName: string,
  values: RemoteSourceFormValues
): Promise<void> {
  const instanceId = get(runtime).instanceId;

  // Emit telemetry
  behaviourEvent?.fireSourceTriggerEvent(
    BehaviourEventAction.SourceAdd,
    BehaviourEventMedium.Button,
    get(appScreen),
    MetricsEventSpace.Modal
  );

  // If project is uninitialized, initialize an empty project
  const isProjectInitialized = await isProjectInitializedV2(
    queryClient,
    instanceId
  );
  if (!isProjectInitialized) {
    await runtimeServiceUnpackEmpty(instanceId, {
      title: EMPTY_PROJECT_TITLE,
    });
  }

  // Convert the form values to Source YAML
  // TODO: Quite a few adhoc code is being added. We should revisit the way we generate the yaml.
  const formValues = Object.fromEntries(
    Object.entries(values).map(([key, value]) => {
      switch (key) {
        case "project_id":
          return [key, value];
        default:
          return [fromYupFriendlyKey(key), value];
      }
    })
  );
  const yaml = compileCreateSourceYAML(formValues, connectorName);

  await createSource(instanceId, values.sourceName, yaml);
  goto(`/source/${values.sourceName}`);
}
