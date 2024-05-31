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
  V1ConnectorDriver,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import {
  compileConnectorYAML,
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "../../connectors/code-utils";
import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "../../entity-management/entity-mappers";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitialized } from "../../welcome/is-project-initialized";
import { compileCreateSourceYAML } from "../sourceUtils";
import { fromYupFriendlyKey } from "./yupSchemas";

export interface RemoteSourceFormValues {
  // sourceName: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: any;
}

export async function submitRemoteSourceForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
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
  const projectInitialized = await isProjectInitialized(
    queryClient,
    instanceId,
  );
  if (!projectInitialized) {
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

  /**
   * Sources
   */

  if (!connector.implementsOlap) {
    // Make a new <source>.yaml file
    await runtimeServicePutFile(instanceId, {
      path: getFileAPIPathFromNameAndType(values.sourceName, EntityType.Table),
      blob: compileCreateSourceYAML(formValues, connector.name as string),
      create: true,
      createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
    });

    await checkSourceImported(
      queryClient,
      getFilePathFromNameAndType(values.sourceName, EntityType.Table),
    );

    return;
  }

  /**
   * Connectors
   */

  // Make a new `<connector>.yaml` file
  await runtimeServicePutFile(instanceId, {
    path: getFileAPIPathFromNameAndType(
      connector.name as string,
      EntityType.Connector,
    ),
    blob: compileConnectorYAML(connector, formValues),
    create: true,
    createOnly: false,
  });

  // Update the `.env` file
  await runtimeServicePutFile(instanceId, {
    path: ".env",
    blob: await updateDotEnvWithSecrets(connector, formValues),
    create: true,
    createOnly: false,
  });

  // Update the `rill.yaml` file
  await runtimeServicePutFile(instanceId, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithOlapConnector(connector.name as string),
    create: true,
    createOnly: false,
  });
}
