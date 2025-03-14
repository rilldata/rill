import { goto, invalidate } from "$app/navigation";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import { behaviourEvent } from "../../../metrics/initMetrics";
import {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "../../../metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
import {
  type V1ConnectorDriver,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceDeleteFile,
  runtimeServiceGetFile,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import {
  compileConnectorYAML,
  deleteEnvVariable,
  makeDotEnvConnectorKey,
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "../../connectors/code-utils";
import { testConnectorConnection } from "../../connectors/olap/test-connection";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { getName } from "../../entity-management/name-utils";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitialized } from "../../welcome/is-project-initialized";
import { compileSourceYAML, maybeRewriteToDuckDb } from "../sourceUtils";

interface AddDataFormValues {
  // name: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: unknown;
}

export async function submitAddSourceForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId);

  const [rewrittenConnector, rewrittenFormValues] = maybeRewriteToDuckDb(
    connector,
    formValues,
  );

  // Make a new <source>.yaml file
  const newSourceFilePath = getFileAPIPathFromNameAndType(
    formValues.name as string,
    EntityType.Table,
  );
  await runtimeServicePutFile(instanceId, {
    path: newSourceFilePath,
    blob: compileSourceYAML(rewrittenConnector, rewrittenFormValues),
    create: true,
    createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
  });

  // Create or update the `.env` file
  await runtimeServicePutFile(instanceId, {
    path: ".env",
    blob: await updateDotEnvWithSecrets(
      queryClient,
      rewrittenConnector,
      rewrittenFormValues,
    ),
    create: true,
    createOnly: false,
  });

  await goto(`/files/${newSourceFilePath}`);
}

export async function submitAddOLAPConnectorForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId);

  const newConnectorName = getName(
    connector.name as string,
    fileArtifacts.getNamesForKind(ResourceKind.Connector),
  );

  // Make a new `<connector>.yaml` file
  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    newConnectorName,
    EntityType.Connector,
  );
  await runtimeServicePutFile(instanceId, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, formValues),
    create: true,
    createOnly: false,
  });

  // Check if .env file exists before we update it
  let envFileExisted = false;
  try {
    const envFile = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    envFileExisted = !!envFile.blob;
  } catch (error) {
    // If file doesn't exist, envFileExisted remains false
    if (!error?.response?.data?.message?.includes("no such file")) {
      throw error; // Re-throw if it's a different error
    }
  }

  // Create or update the `.env` file
  const newEnvBlob = await updateDotEnvWithSecrets(
    queryClient,
    connector,
    formValues,
  );
  await runtimeServicePutFile(instanceId, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Test the connection
  const result = await testConnectorConnection(
    instanceId,
    newConnectorFilePath,
    newConnectorName,
  );

  // If the connection test fails, clean-up the files
  if (!result.success) {
    // Clean-up the `connector.yaml` file
    await runtimeServiceDeleteFile(instanceId, {
      path: newConnectorFilePath,
    });

    // Clean-up the `.env` file
    if (!envFileExisted) {
      // If .env file didn't exist before, delete it
      await runtimeServiceDeleteFile(instanceId, {
        path: ".env",
      });
    } else {
      // If .env file existed before, remove only the secrets we added
      const secretKeys =
        connector.configProperties
          ?.filter((property) => property.secret)
          .map((property) => property.key) || [];

      let updatedEnvBlob = newEnvBlob;
      for (const key of secretKeys) {
        if (key) {
          const envKey = makeDotEnvConnectorKey(connector.name as string, key);
          updatedEnvBlob = deleteEnvVariable(updatedEnvBlob, envKey);
        }
      }

      await runtimeServicePutFile(instanceId, {
        path: ".env",
        blob: updatedEnvBlob,
        create: true,
        createOnly: false,
      });
    }

    throw new Error(result.error || "Unable to establish a connection");
  }

  // The connection test passed

  // Update the `rill.yaml` file
  await runtimeServicePutFile(instanceId, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithOlapConnector(queryClient, newConnectorName),
    create: true,
    createOnly: false,
  });

  // Go to the new connector file
  await goto(`/files/${newConnectorFilePath}`);
}

async function beforeSubmitForm(instanceId: string) {
  // Emit telemetry
  await behaviourEvent?.fireSourceTriggerEvent(
    BehaviourEventAction.SourceAdd,
    BehaviourEventMedium.Button,
    getScreenNameFromPage(),
    MetricsEventSpace.Modal,
  );

  // If project is uninitialized, initialize an empty project
  const projectInitialized = await isProjectInitialized(instanceId);
  if (!projectInitialized) {
    await runtimeServiceUnpackEmpty(instanceId, {
      displayName: EMPTY_PROJECT_TITLE,
    });

    // Race condition: invalidate("init") must be called before we navigate to
    // `/files/${newFilePath}`. invalidate("init") is also called in the
    // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
    await invalidate("init");
  }
}
