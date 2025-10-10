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
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "../../connectors/code-utils";
import {
  runtimeServicePutFileAndWaitForReconciliation,
  waitForResourceReconciliation,
} from "../../entity-management/actions";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { getName } from "../../entity-management/name-utils";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitialized } from "../../welcome/is-project-initialized";
import { compileSourceYAML, prepareSourceFormData } from "../sourceUtils";
import { OLAP_ENGINES } from "./constants";

interface AddDataFormValues {
  // name: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: unknown;
}

async function beforeSubmitForm(
  instanceId: string,
  connector?: V1ConnectorDriver,
) {
  // Emit telemetry
  behaviourEvent?.fireSourceTriggerEvent(
    BehaviourEventAction.SourceAdd,
    BehaviourEventMedium.Button,
    getScreenNameFromPage(),
    MetricsEventSpace.Modal,
  );

  // If project is uninitialized, initialize an empty project
  const projectInitialized = await isProjectInitialized(instanceId);
  if (!projectInitialized) {
    // Determine the OLAP engine based on the connector being added
    let olapEngine = "duckdb"; // Default for data sources

    if (connector && OLAP_ENGINES.includes(connector.name as string)) {
      // For OLAP engines, use the connector name as the OLAP engine
      olapEngine = connector.name as string;
    }

    await runtimeServiceUnpackEmpty(instanceId, {
      displayName: EMPTY_PROJECT_TITLE,
      olap: olapEngine, // Explicitly set OLAP based on connector type
    });

    // Race condition: invalidate("init") must be called before we navigate to
    // `/files/${newFilePath}`. invalidate("init") is also called in the
    // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
    await invalidate("init");
  }
}

async function rollbackChanges(
  instanceId: string,
  newFilePath: string,
  originalEnvBlob: string | undefined,
) {
  // Clean-up the file
  await runtimeServiceDeleteFile(instanceId, {
    path: newFilePath,
  });

  // Clean-up the `.env` file
  if (!originalEnvBlob) {
    // If .env file didn't exist before, delete it
    await runtimeServiceDeleteFile(instanceId, {
      path: ".env",
    });
  } else {
    // If .env file existed before, restore its original content
    await runtimeServicePutFile(instanceId, {
      path: ".env",
      blob: originalEnvBlob,
      create: true,
      createOnly: false,
    });
  }
}

async function setOlapConnectorInRillYAML(
  queryClient: QueryClient,
  instanceId: string,
  newConnectorName: string,
): Promise<void> {
  await runtimeServicePutFile(instanceId, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithOlapConnector(queryClient, newConnectorName),
    create: true,
    createOnly: false,
  });
}

// Check for an existing `.env` file
// Store the original `.env` blob so we can restore it in case of errors
async function getOriginalEnvBlob(
  queryClient: QueryClient,
  instanceId: string,
): Promise<string | undefined> {
  try {
    const envFile = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    return envFile.blob;
  } catch (error) {
    const fileNotFound =
      error?.response?.data?.message?.includes("no such file");
    if (fileNotFound) {
      // Do nothing. We'll create the `.env` file below.
      return undefined;
    } else {
      // We have a problem. Throw the error.
      throw error;
    }
  }
}

export async function submitAddSourceForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  saveAnyway: boolean = false,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId, connector);

  const newSourceName = formValues.name as string;

  const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
    connector,
    formValues,
  );

  // Make a new <source>.yaml file
  const newSourceFilePath = getFileAPIPathFromNameAndType(
    newSourceName,
    EntityType.Table,
  );
  await runtimeServicePutFile(instanceId, {
    path: newSourceFilePath,
    blob: compileSourceYAML(rewrittenConnector, rewrittenFormValues),
    create: true,
    createOnly: false,
  });

  const originalEnvBlob = await getOriginalEnvBlob(queryClient, instanceId);

  // Create or update the `.env` file
  const newEnvBlob = await updateDotEnvWithSecrets(
    queryClient,
    rewrittenConnector,
    rewrittenFormValues,
    "source",
  );

  // Make sure the file has reconciled before testing the connection
  await runtimeServicePutFileAndWaitForReconciliation(instanceId, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Wait for source resource-level reconciliation
  // This must happen after .env reconciliation since sources depend on secrets
  try {
    await waitForResourceReconciliation(
      instanceId,
      newSourceName,
      ResourceKind.Model,
      connector.name as string,
    );
  } catch (error) {
    if (!saveAnyway) {
      // The source file was already created, so we need to delete it
      await rollbackChanges(instanceId, newSourceFilePath, originalEnvBlob);
      const errorDetails = (error as any).details;

      throw {
        message: error.message || "Unable to establish a connection",
        details:
          errorDetails && errorDetails !== error.message
            ? errorDetails
            : undefined,
      };
    }
    // If saveAnyway is true, we continue despite the reconciliation error
    // The file will be saved but may have connection issues
  }

  // Check for file errors
  // If the model file has errors, rollback the changes
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    instanceId,
    newSourceFilePath,
  );
  if (errorMessage && !saveAnyway) {
    await rollbackChanges(instanceId, newSourceFilePath, originalEnvBlob);
    throw new Error(errorMessage);
  }
  // If saveAnyway is true, we continue despite file errors
  // The file will be saved but may have parsing issues

  await goto(`/files/${newSourceFilePath}`);
}

export async function submitAddConnectorForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  saveAnyway: boolean = false,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId, connector);

  const newConnectorName = getName(
    connector.name as string,
    fileArtifacts.getNamesForKind(ResourceKind.Connector),
  );

  /**
   * Optimistic updates:
   * 1. Make a new `<connector>.yaml` file
   * 2. Create/update the `.env` file with connector secrets
   */

  // Make a new `<connector>.yaml` file
  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    newConnectorName,
    EntityType.Connector,
  );
  await runtimeServicePutFile(instanceId, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, formValues, {
      connectorInstanceName: newConnectorName,
    }),
    create: true,
    createOnly: false,
  });

  const originalEnvBlob = await getOriginalEnvBlob(queryClient, instanceId);

  // Create or update the `.env` file
  const newEnvBlob = await updateDotEnvWithSecrets(
    queryClient,
    connector,
    formValues,
    "connector",
    newConnectorName,
  );

  // Make sure the file has reconciled before testing the connection
  await runtimeServicePutFileAndWaitForReconciliation(instanceId, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Wait for connector resource-level reconciliation
  // This must happen after .env reconciliation since connectors depend on secrets
  try {
    await waitForResourceReconciliation(
      instanceId,
      newConnectorName,
      ResourceKind.Connector,
      connector.name as string,
    );
  } catch (error) {
    if (!saveAnyway) {
      // The connector file was already created, so we need to delete it
      await rollbackChanges(instanceId, newConnectorFilePath, originalEnvBlob);
      const errorDetails = (error as any).details;

      throw {
        message: error.message || "Unable to establish a connection",
        details:
          errorDetails && errorDetails !== error.message
            ? errorDetails
            : undefined,
      };
    }
    // If saveAnyway is true, we continue despite the reconciliation error
    // The file will be saved but may have connection issues
  }

  // Check for file errors
  // If the connector file has errors, rollback the changes
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    instanceId,
    newConnectorFilePath,
  );
  if (errorMessage && !saveAnyway) {
    await rollbackChanges(instanceId, newConnectorFilePath, originalEnvBlob);
    throw new Error(errorMessage);
  }
  // If saveAnyway is true, we continue despite file errors
  // The file will be saved but may have parsing issues

  if (OLAP_ENGINES.includes(connector.name as string)) {
    await setOlapConnectorInRillYAML(queryClient, instanceId, newConnectorName);
  }

  // Go to the new connector file
  await goto(`/files/${newConnectorFilePath}`);
}
