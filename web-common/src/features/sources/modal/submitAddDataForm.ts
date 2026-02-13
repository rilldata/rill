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
  updateRillYAMLWithAiConnector,
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
import { AI_CONNECTORS, OLAP_ENGINES } from "./constants";
import { getConnectorSchema } from "./connector-schemas";
import {
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "../../templates/schema-utils";

interface AddDataFormValues {
  // name: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: unknown;
}

// Track connector file paths that were created via Save Anyway so
// in-flight Test-and-Connect submissions don't roll them back.
const savedAnywayPaths = new Set<string>();

const connectorSubmissions = new Map<
  string,
  {
    promise: Promise<string>;
    connectorName: string;
  }
>();

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

async function setAiConnectorInRillYAML(
  queryClient: QueryClient,
  instanceId: string,
  newConnectorName: string,
): Promise<void> {
  await runtimeServicePutFile(instanceId, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithAiConnector(queryClient, newConnectorName),
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

async function saveConnectorAnyway(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  newConnectorName: string,
  instanceId?: string,
): Promise<void> {
  const resolvedInstanceId = instanceId ?? get(runtime).instanceId;
  const schema = getConnectorSchema(connector.name ?? "");
  const schemaFields = schema
    ? getSchemaFieldMetaList(schema, { step: "connector" })
    : [];
  const schemaSecretKeys = schema
    ? getSchemaSecretKeys(schema, { step: "connector" })
    : [];
  const schemaStringKeys = schema
    ? getSchemaStringKeys(schema, { step: "connector" })
    : [];

  // Create connector file
  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    newConnectorName,
    EntityType.Connector,
  );

  // Mark to avoid rollback by concurrent submissions
  savedAnywayPaths.add(newConnectorFilePath);

  // Update .env file with secrets (keep ordering consistent with Test and Connect)
  const newEnvBlob = await updateDotEnvWithSecrets(
    queryClient,
    connector,
    formValues,
    "connector",
    newConnectorName,
    { secretKeys: schemaSecretKeys },
  );

  await runtimeServicePutFile(resolvedInstanceId, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Always create/overwrite to ensure the connector file is created immediately
  await runtimeServicePutFile(resolvedInstanceId, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, formValues, {
      connectorInstanceName: newConnectorName,
      orderedProperties: schemaFields,
      secretKeys: schemaSecretKeys,
      stringKeys: schemaStringKeys,
      fieldFilter: schemaFields
        ? (property) => !("internal" in property && property.internal)
        : undefined,
    }),
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

  if (AI_CONNECTORS.includes(connector.name as string)) {
    await setAiConnectorInRillYAML(
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
  const schema = getConnectorSchema(connector.name ?? "");
  const schemaFields = schema
    ? getSchemaFieldMetaList(schema, { step: "connector" })
    : [];
  const schemaSecretKeys = schema
    ? getSchemaSecretKeys(schema, { step: "connector" })
    : [];
  const schemaStringKeys = schema
    ? getSchemaStringKeys(schema, { step: "connector" })
    : [];

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
    }

    // If Test and Connect is clicked while another operation is running,
    // wait for it to complete
    return existingSubmission.promise;
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

    let originalEnvBlob: string | undefined;
    let envWritten = false;
    let connectorCreated = false;

    try {
      // Check if operation was aborted
      if (abortController.signal.aborted) {
        throw new Error("Operation cancelled");
      }

      // Capture original .env and compute updated contents up front
      originalEnvBlob = await getOriginalEnvBlob(queryClient, instanceId);
      const newEnvBlob = await updateDotEnvWithSecrets(
        queryClient,
        connector,
        formValues,
        "connector",
        newConnectorName,
        { secretKeys: schemaSecretKeys },
      );

      if (saveAnyway) {
        // Save Anyway: bypass reconciliation entirely via centralized helper
        await saveConnectorAnyway(
          queryClient,
          connector,
          formValues,
          newConnectorName,
          instanceId,
        );
        return newConnectorName;
      }

      /**
       * Optimistic updates (Test and Connect):
       * 1. Write the `.env` file so secrets exist before connector reconciliation
       * 2. Create the `<connector>.yaml` file
       * 3. Wait for reconciliation and surface any file errors
       */
      await runtimeServicePutFileAndWaitForReconciliation(instanceId, {
        path: ".env",
        blob: newEnvBlob,
        create: true,
        createOnly: false,
      });
      envWritten = true;

      await runtimeServicePutFile(
        instanceId,
        {
          path: newConnectorFilePath,
          blob: compileConnectorYAML(connector, formValues, {
            connectorInstanceName: newConnectorName,
            orderedProperties: schemaFields,
            secretKeys: schemaSecretKeys,
            stringKeys: schemaStringKeys,
            fieldFilter: schemaFields
              ? (property) => !("internal" in property && property.internal)
              : undefined,
          }),
          create: true,
          createOnly: false,
        },
        abortController.signal,
      );
      connectorCreated = true;

      // Wait for connector resource-level reconciliation
      // This must happen after .env reconciliation since connectors depend on secrets
      await waitForResourceReconciliation(
        instanceId,
        newConnectorName,
        ResourceKind.Connector,
      );

      // Check for file errors
      // If the connector file has errors, rollback the changes
      const errorMessage = await fileArtifacts.checkFileErrors(
        queryClient,
        instanceId,
        newConnectorFilePath,
      );
      if (errorMessage) {
        throw new Error(errorMessage);
      }

      if (OLAP_ENGINES.includes(connector.name as string)) {
        await setOlapConnectorInRillYAML(
          queryClient,
          instanceId,
          newConnectorName,
        );
      }

      // Note: Currently unreachable for AI connectors (they always use
      // saveAnyway via isAiConnector), but kept as a safety net in case
      // the flow changes to allow AI connectors through the test path.
      if (AI_CONNECTORS.includes(connector.name as string)) {
        await setAiConnectorInRillYAML(
          queryClient,
          instanceId,
          newConnectorName,
        );
      }

      // Go to the new connector file
      await goto(`/files/${newConnectorFilePath}`);
      return newConnectorName;
    } catch (error) {
      // If the operation was aborted, don't treat it as an error
      if (abortController.signal.aborted) {
        console.log("Operation was cancelled");
        return newConnectorName;
      }

      const shouldRollbackConnectorFile =
        !savedAnywayPaths.has(newConnectorFilePath) &&
        (envWritten || connectorCreated);

      if (shouldRollbackConnectorFile) {
        await rollbackChanges(
          instanceId,
          newConnectorFilePath,
          originalEnvBlob,
        );
      }

      const errorDetails = (error as any).details;
      if (errorDetails && errorDetails !== (error as any).message) {
        throw {
          message: (error as any).message || "Unable to establish a connection",
          details: errorDetails,
        };
      }

      throw error;
    } finally {
      // Mark the submission as completed but keep the connector name around
      // so a subsequent "Save Anyway" can still reuse the same connector file
      connectorSubmissions.delete(uniqueConnectorSubmissionKey);
    }
  })();

  // Store the submission promise
  connectorSubmissions.set(uniqueConnectorSubmissionKey, {
    promise: submissionPromise,
    connectorName: newConnectorName,
  });

  // Wait for the submission to complete
  const resolvedConnectorName = await submissionPromise;
  return resolvedConnectorName;
}

export async function submitAddSourceForm(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  connectorInstanceName?: string,
): Promise<void> {
  const instanceId = get(runtime).instanceId;
  await beforeSubmitForm(instanceId, connector);
  const newSourceName = formValues.name as string;

  const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
    connector,
    formValues,
    { connectorInstanceName },
  );
  const schema = getConnectorSchema(rewrittenConnector.name ?? "");
  const schemaSecretKeys = schema
    ? getSchemaSecretKeys(schema, { step: "source" })
    : [];
  const schemaStringKeys = schema
    ? getSchemaStringKeys(schema, { step: "source" })
    : [];

  // When connector is rewritten to DuckDB (e.g., S3 -> DuckDB), don't use
  // the original connectorInstanceName in YAML. The original connector is
  // referenced via create_secrets_from_connectors for credential access.
  const isRewrittenToDuckDb =
    rewrittenConnector.name === "duckdb" && connector.name !== "duckdb";
  const yamlConnectorInstanceName = isRewrittenToDuckDb
    ? undefined
    : connectorInstanceName;

  // Make a new <source>.yaml file
  const newSourceFilePath = getFileAPIPathFromNameAndType(
    newSourceName,
    EntityType.Table,
  );
  await runtimeServicePutFile(instanceId, {
    path: newSourceFilePath,
    blob: compileSourceYAML(rewrittenConnector, rewrittenFormValues, {
      secretKeys: schemaSecretKeys,
      stringKeys: schemaStringKeys,
      connectorInstanceName: yamlConnectorInstanceName,
      originalDriverName: connector.name || undefined,
    }),
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
    undefined,
    { secretKeys: schemaSecretKeys },
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
