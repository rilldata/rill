import { goto } from "$app/navigation";
import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import { appScreen } from "../../../layout/app-store";
import { behaviourEvent } from "../../../metrics/initMetrics";
import {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "../../../metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
import { connectorToSourceConnectionType } from "../../../metrics/service/SourceEventTypes";
import {
  runtimeServicePutFileAndReconcile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { invalidateAfterReconcile } from "../../../runtime-client/invalidation";
import { runtime } from "../../../runtime-client/runtime-store";
import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifactsStore } from "../../entity-management/file-artifacts-store";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitializedV2 } from "../../welcome/is-project-initialized";
import { createModelFromSourceV2 } from "../createModel";
import {
  compileCreateSourceYAML,
  emitSourceErrorTelemetry,
  emitSourceSuccessTelemetry,
  getSourceError,
} from "../sourceUtils";
import { fromYupFriendlyKey } from "./yupSchemas";

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

  // Attempt to create & import the source
  const resp = await runtimeServicePutFileAndReconcile({
    instanceId,
    path: getFilePathFromNameAndType(values.sourceName, EntityType.Table),
    blob: yaml,
    create: true,
    createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
    dry: false,
    strict: false,
  });
  invalidateAfterReconcile(queryClient, instanceId, resp);
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  // Navigate according to failure or success
  const hasSourceYAMLErrors = resp.errors.length > 0;
  if (hasSourceYAMLErrors) {
    // Error

    // Emit telemetry
    const sourceError = getSourceError(resp.errors, values.sourceName);
    emitSourceErrorTelemetry(
      MetricsEventSpace.Modal,
      get(appScreen),
      sourceError?.message,
      connectorToSourceConnectionType[connectorName],
      formValues?.uri
    );

    // Show the source YAML editor
    goto(`/source/${values.sourceName}`);
  } else {
    // Success

    // Emit telemetry
    emitSourceSuccessTelemetry(
      MetricsEventSpace.Modal,
      get(appScreen),
      BehaviourEventMedium.Button,
      connectorToSourceConnectionType[connectorName],
      formValues?.uri
    );

    // Create and navigate to a `select *` model
    createModelFromSourceV2(queryClient, values.sourceName);
  }
}
