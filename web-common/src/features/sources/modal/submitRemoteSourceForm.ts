import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import { behaviourEvent } from "../../../metrics/initMetrics";
import {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "../../../metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
import {
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "../../entity-management/entity-mappers";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitializedV2 } from "../../welcome/is-project-initialized";
import { compileCreateSourceYAML } from "../sourceUtils";
import { fromYupFriendlyKey } from "./yupSchemas";

export interface RemoteSourceFormValues {
  sourceName: string;
  [key: string]: any;
}

export async function submitRemoteSourceForm(
  queryClient: QueryClient,
  connectorName: string,
  values: RemoteSourceFormValues,
): Promise<void> {
  const instanceId = get(runtime).instanceId;

  // Emit telemetry
  await behaviourEvent?.fireSourceTriggerEvent(
    BehaviourEventAction.SourceAdd,
    BehaviourEventMedium.Button,
    getScreenNameFromPage(),
    MetricsEventSpace.Modal,
  );

  // If project is uninitialized, initialize an empty project
  const isProjectInitialized = await isProjectInitializedV2(
    queryClient,
    instanceId,
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
        case "account":
        case "output_location":
        case "workgroup":
        case "database_url":
          return [key, value];
        default:
          return [fromYupFriendlyKey(key), value];
      }
    }),
  );
  const yaml = compileCreateSourceYAML(formValues, connectorName);

  // Attempt to create & import the source
  await runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(values.sourceName, EntityType.Table),
    {
      blob: yaml,
      create: true,
      createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
    },
  );
  checkSourceImported(
    queryClient,
    getFilePathFromNameAndType(values.sourceName, EntityType.Table),
  );

  // TODO: telemetry
  // Emit telemetry
  // const hasSourceYAMLErrors = resp.errors.length > 0;
  // if (hasSourceYAMLErrors) {
  //   // Error
  //   const sourceError = getSourceError(resp.errors, values.sourceName);
  //   emitSourceErrorTelemetry(
  //     MetricsEventSpace.Modal,
  //     get(appScreen),
  //     sourceError?.message,
  //     connectorToSourceConnectionType[connectorName],
  //     formValues?.uri || ""
  //   );
  // } else {
  //   // Success
  //   emitSourceSuccessTelemetry(
  //     MetricsEventSpace.Modal,
  //     get(appScreen),
  //     BehaviourEventMedium.Button,
  //     connectorToSourceConnectionType[connectorName],
  //     formValues?.uri || ""
  //   );
  // }
}
