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

// FIXME: consolidate this
// Source YAML - `type: model`
export async function submitAddSourceForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId);

  const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
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
    createOnly: false,
  });

  // Check for an existing `.env` file
  // Store the original `.env` blob so we can restore it in case of errors
  let originalEnvBlob: string | undefined;
  try {
    const envFile = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    originalEnvBlob = envFile.blob;
  } catch (error) {
    const fileNotFound =
      error?.response?.data?.message?.includes("no such file");
    if (fileNotFound) {
      // Do nothing. We'll create the `.env` file below.
    } else {
      // We have a problem. Throw the error.
      throw error;
    }
  }

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

  // Wait for source resource reconciliation
  // This must happen after .env reconciliation since sources depend on secrets
  try {
    await waitForResourceReconciliation(
      instanceId,
      formValues.name as string,
      ResourceKind.Model,
      connector.name as string,
    );
  } catch (error) {
    // The source file was already created, so we need to delete it
    await rollbackSourceChanges(instanceId, newSourceFilePath, originalEnvBlob);
    const errorDetails = (error as any).details;

    // Provide more helpful error messages for specific connectors
    let errorMessage = error.message || "Unable to establish a connection";
    if (
      errorMessage.includes(
        "Resource configuration failed to reconcile and was automatically deleted",
      )
    ) {
      if (connector.name === "gcs") {
        errorMessage =
          "GCS connection failed. Please check your credentials and bucket permissions.";
      } else if (connector.name === "s3") {
        errorMessage =
          "S3 connection failed. Please check your credentials and bucket permissions.";
      } else {
        errorMessage = `${connector.name} connection failed. Please check your connection details and credentials.`;
      }
    }

    throw {
      message: errorMessage,
      details:
        errorDetails && errorDetails !== error.message
          ? errorDetails
          : undefined,
    };
  }

  // Check for file errors
  // If the source file has errors, rollback the changes
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    instanceId,
    newSourceFilePath,
  );
  if (errorMessage) {
    await rollbackSourceChanges(instanceId, newSourceFilePath, originalEnvBlob);
    throw new Error(errorMessage);
  }

  await goto(`/files/${newSourceFilePath}`);
}

// Connector YAML - `type: connector`
export async function submitAddConnectorForm(
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

  // Check for an existing `.env` file
  // Store the original `.env` blob so we can restore it in case of errors
  let originalEnvBlob: string | undefined;
  try {
    const envFile = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    originalEnvBlob = envFile.blob;
  } catch (error) {
    const fileNotFound =
      error?.response?.data?.message?.includes("no such file");
    if (fileNotFound) {
      // Do nothing. We'll create the `.env` file below.
    } else {
      // We have a problem. Throw the error.
      throw error;
    }
  }

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
    // The connector file was already created, so we need to delete it
    await runtimeServiceDeleteFile(instanceId, {
      path: newConnectorFilePath,
    });
    throw {
      message: error.message || "Unable to establish a connection",
      details: (error as any).details || error.message,
    };
  }
  // }

  // Check for file errors
  // If the connector file has errors, rollback the changes
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    instanceId,
    newConnectorFilePath,
  );
  if (errorMessage) {
    await rollbackConnectorChanges(
      instanceId,
      newConnectorFilePath,
      originalEnvBlob,
    );
    throw new Error(errorMessage);
  }

  // Connection testing is now handled by resource reconciliation above
  // No need for separate OLAP-specific testing since reconciliation includes Ping() calls

  /**
   * Connection successful: Complete the setup
   * Update the project configuration and navigate to the new connector
   */

  // Update the `rill.yaml` file only for actual OLAP connectors (ClickHouse, Druid, Pinot)
  if (OLAP_ENGINES.includes(connector.name as string)) {
    await runtimeServicePutFile(instanceId, {
      path: "rill.yaml",
      blob: await updateRillYAMLWithOlapConnector(
        queryClient,
        newConnectorName,
      ),
      create: true,
      createOnly: false,
    });
  }

  // Go to the new connector file
  await goto(`/files/${newConnectorFilePath}`);
}

async function beforeSubmitForm(instanceId: string) {
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
    await runtimeServiceUnpackEmpty(instanceId, {
      displayName: EMPTY_PROJECT_TITLE,
    });

    // Race condition: invalidate("init") must be called before we navigate to
    // `/files/${newFilePath}`. invalidate("init") is also called in the
    // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
    await invalidate("init");
  }
}

// FIXME: consolidate this
async function rollbackConnectorChanges(
  instanceId: string,
  newConnectorFilePath: string,
  originalEnvBlob: string | undefined,
) {
  // Clean-up the `connector.yaml` file
  await runtimeServiceDeleteFile(instanceId, {
    path: newConnectorFilePath,
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

async function rollbackSourceChanges(
  instanceId: string,
  newSourceFilePath: string,
  originalEnvBlob: string | undefined,
) {
  // Clean-up the `source.yaml` file
  await runtimeServiceDeleteFile(instanceId, {
    path: newSourceFilePath,
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
