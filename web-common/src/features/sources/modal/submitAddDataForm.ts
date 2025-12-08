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
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceAnalyzeConnectors,
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
import { connectorStepStore } from "./connectorStepStore";
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
import { compileModelYAML, prepareSourceFormData } from "../sourceUtils";
import { CONNECTORS_USING_INSTANCE_SECRETS, OLAP_ENGINES } from "./constants";

interface AddDataFormValues {
  // name: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: unknown;
}

// Track connector file paths that were created via Save Anyway so
// in-flight Test-and-Connect submissions don't roll them back.
const savedAnywayPaths = new Set<string>();

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

/**
 * Resolves the connector instance name for a given driver by analyzing available connectors.
 *
 * Looks up all connectors in the project and finds those matching the specified driver name.
 * Prefers a connector instance that is literally named after the driver (e.g., "s3" for the S3 driver),
 * otherwise returns the first matching connector instance name. Returns undefined if no matching
 * connectors are found or if the lookup fails.
 */
async function resolveConnectorInstanceName(
  queryClient: QueryClient,
  instanceId: string,
  driverName: string,
): Promise<string | undefined> {
  try {
    const analyzeConnectorsQueryKey =
      getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId);
    const analyzeConnectorsQueryFn = async () =>
      runtimeServiceAnalyzeConnectors(instanceId);
    const connectors = await queryClient.fetchQuery({
      queryKey: analyzeConnectorsQueryKey,
      queryFn: analyzeConnectorsQueryFn,
    });

    const matchingConnectorNames =
      connectors?.connectors
        ?.filter((c) => c.driver?.name === driverName)
        .map((c) => c.name)
        .filter(Boolean) ?? [];

    if (matchingConnectorNames.length === 0) return undefined;

    // Prefer an instance literally named after the driver (e.g., "s3" or "azure") if present,
    // otherwise pick the first one.
    const preferred =
      matchingConnectorNames.find((name) => name === driverName) ??
      matchingConnectorNames[0];

    return preferred as string;
  } catch {
    // If lookup fails, return undefined and rely on auto-detection.
    return undefined;
  }
}

export async function submitAddSourceForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId, connector);

  const newSourceName = formValues.name as string;

  const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
    connector,
    formValues,
  );

  let connectorInstanceName: string | undefined;
  const stepState = get(connectorStepStore);
  if (stepState?.step === "source") {
    connectorInstanceName = stepState.connectorInstanceName || undefined;
  }
  const connectorName = connector.name as string;

  // Resolve the connector instance name(s) for create_secrets_from_connectors.
  // For supported remote storage sources (e.g., S3, Azure), look up available
  // connectors and use their instance names.
  if (
    !connectorInstanceName &&
    CONNECTORS_USING_INSTANCE_SECRETS.includes(connectorName)
  ) {
    connectorInstanceName = await resolveConnectorInstanceName(
      queryClient,
      instanceId,
      connectorName,
    );
  }

  // Make a new <source>.yaml file
  const newSourceFilePath = getFileAPIPathFromNameAndType(
    newSourceName,
    EntityType.Table,
  );
  await runtimeServicePutFile(instanceId, {
    path: newSourceFilePath,
    blob: compileModelYAML(
      rewrittenConnector.name as string,
      rewrittenFormValues,
      rewrittenConnector.sourceProperties,
      { connectorInstanceName },
    ),
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
    );
  } catch (error) {
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

  // Check for file errors
  // If the model file has errors, rollback the changes
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    instanceId,
    newSourceFilePath,
  );
  if (errorMessage) {
    await rollbackChanges(instanceId, newSourceFilePath, originalEnvBlob);
    throw new Error(errorMessage);
  }

  await goto(`/files/${newSourceFilePath}`);
}

const connectorSubmissions = new Map<
  string,
  {
    promise: Promise<void>;
    connectorName: string;
    completed: boolean;
  }
>();

async function saveConnectorAnyway(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  newConnectorName: string,
  instanceId?: string,
): Promise<void> {
  const resolvedInstanceId = instanceId ?? get(runtime).instanceId;

  // Create connector file
  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    newConnectorName,
    EntityType.Connector,
  );

  // Mark to avoid rollback by concurrent submissions
  savedAnywayPaths.add(newConnectorFilePath);

  // Always create/overwrite to ensure the file is created immediately
  await runtimeServicePutFile(resolvedInstanceId, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, formValues, {
      connectorInstanceName: newConnectorName,
    }),
    create: true,
    createOnly: false,
  });

  // Update .env file with secrets
  const newEnvBlob = await updateDotEnvWithSecrets(
    queryClient,
    connector,
    formValues,
    "connector",
    newConnectorName,
  );

  await runtimeServicePutFile(resolvedInstanceId, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  if (OLAP_ENGINES.includes(connector.name as string)) {
    await setOlapConnectorInRillYAML(
      queryClient,
      resolvedInstanceId,
      newConnectorName,
    );
  }

  // Go to the new connector file
  await goto(`/files/${newConnectorFilePath}`);
}

export async function submitAddConnectorForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  saveAnyway: boolean = false,
): Promise<string> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId, connector);

  // Create a unique key for this connector submission
  const uniqueConnectorSubmissionKey = `${instanceId}:${connector.name}`;

  const newConnectorName = getName(
    connector.name as string,
    fileArtifacts.getNamesForKind(ResourceKind.Connector),
  );

  // Check if there's already an ongoing submission for this connector
  const existingSubmission = connectorSubmissions.get(
    uniqueConnectorSubmissionKey,
  );

  if (existingSubmission) {
    if (saveAnyway) {
      // If Save Anyway is clicked while Test and Connect is running,
      // proceed immediately without waiting for the ongoing operation
      // Clean up the existing submission
      connectorSubmissions.delete(uniqueConnectorSubmissionKey);

      // Use the same connector name from the ongoing operation
      const newConnectorName = existingSubmission.connectorName;

      // Proceed immediately with Save Anyway logic
      await saveConnectorAnyway(
        queryClient,
        connector,
        formValues,
        newConnectorName,
        instanceId,
      );
      return newConnectorName;
    } else if (!existingSubmission.completed) {
      // If Test and Connect is clicked while another operation is running,
      // wait for it to complete
      await existingSubmission.promise;
      return existingSubmission.connectorName;
    }
  }

  // Create abort controller for this submission
  const abortController = new AbortController();

  // Create a new submission promise
  const submissionPromise = (async () => {
    // Create connector file path outside try block for cleanup
    const newConnectorFilePath = getFileAPIPathFromNameAndType(
      newConnectorName,
      EntityType.Connector,
    );

    try {
      // Check if operation was aborted
      if (abortController.signal.aborted) {
        throw new Error("Operation cancelled");
      }
      /**
       * Optimistic updates:
       * 1. Make a new `<connector>.yaml` file
       * 2. Create/update the `.env` file with connector secrets
       */

      // Make a new `<connector>.yaml` file
      if (saveAnyway) {
        // Save Anyway: bypass reconciliation entirely via centralized helper
        await saveConnectorAnyway(
          queryClient,
          connector,
          formValues,
          newConnectorName,
          instanceId,
        );
        return;
      } else {
        // For Test and Connect, create file normally with abort signal
        await runtimeServicePutFile(
          instanceId,
          {
            path: newConnectorFilePath,
            blob: compileConnectorYAML(connector, formValues, {
              connectorInstanceName: newConnectorName,
            }),
            create: true,
            createOnly: false,
          },
          abortController.signal,
        );
      }

      const originalEnvBlob = await getOriginalEnvBlob(queryClient, instanceId);

      // Create or update the `.env` file
      const newEnvBlob = await updateDotEnvWithSecrets(
        queryClient,
        connector,
        formValues,
        "connector",
        newConnectorName,
      );

      if (!saveAnyway) {
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
          );
        } catch (error) {
          // The connector file was already created, so we would delete it
          // unless Save Anyway has already created it intentionally.
          if (!savedAnywayPaths.has(newConnectorFilePath)) {
            await rollbackChanges(
              instanceId,
              newConnectorFilePath,
              originalEnvBlob,
            );
          }
          const errorDetails = (error as any).details;

          throw {
            message: error.message || "Unable to establish a connection",
            details:
              errorDetails && errorDetails !== error.message
                ? errorDetails
                : undefined,
          };
        }

        // Check for file errors
        // If the connector file has errors, rollback the changes
        const errorMessage = await fileArtifacts.checkFileErrors(
          queryClient,
          instanceId,
          newConnectorFilePath,
        );
        if (errorMessage) {
          if (!savedAnywayPaths.has(newConnectorFilePath)) {
            await rollbackChanges(
              instanceId,
              newConnectorFilePath,
              originalEnvBlob,
            );
          }
          throw new Error(errorMessage);
        }
      }

      if (OLAP_ENGINES.includes(connector.name as string)) {
        await setOlapConnectorInRillYAML(
          queryClient,
          instanceId,
          newConnectorName,
        );
      }

      // Go to the new connector file
      await goto(`/files/${newConnectorFilePath}`);
    } catch (error) {
      // If the operation was aborted, don't treat it as an error
      if (abortController.signal.aborted) {
        console.log("Operation was cancelled");
        return;
      }
      throw error;
    } finally {
      // Mark the submission as completed but keep the connector name around
      // so a subsequent "Save Anyway" can still reuse the same connector file
      const submission = connectorSubmissions.get(uniqueConnectorSubmissionKey);
      if (submission) {
        submission.completed = true;
      }
    }
  })();

  // Store the submission promise
  connectorSubmissions.set(uniqueConnectorSubmissionKey, {
    promise: submissionPromise,
    connectorName: newConnectorName,
    completed: false,
  });

  // Wait for the submission to complete
  await submissionPromise;
  return newConnectorName;
}
