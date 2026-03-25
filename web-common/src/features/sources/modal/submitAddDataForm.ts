import { goto, invalidate } from "$app/navigation";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import type { QueryClient } from "@tanstack/query-core";
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
  runtimeServiceGetResource,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import type { RuntimeClient } from "../../../runtime-client/v2";
import {
  compileConnectorYAML,
  getGenericEnvVarName,
  replaceOrAddEnvVariable,
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
import { sourceIngestionTracker } from "../sources-store";
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

// Track connector file paths that were created via Save (without test) so
// in-flight Test and Connect submissions don't roll them back.
const savedWithoutTestPaths = new Set<string>();

const connectorSubmissions = new Map<
  string,
  {
    promise: Promise<string>;
    connectorName: string;
  }
>();

export async function beforeSubmitForm(
  client: RuntimeClient,
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
  const projectInitialized = await isProjectInitialized(client);
  if (!projectInitialized) {
    // Determine the OLAP engine based on the connector being added
    let olapEngine = "duckdb"; // Default for data sources

    if (connector && OLAP_ENGINES.includes(connector.name as string)) {
      // For OLAP engines, use the connector name as the OLAP engine
      olapEngine = connector.name as string;
    }

    await runtimeServiceUnpackEmpty(client, {
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
  client: RuntimeClient,
  newFilePath: string,
  originalEnvBlob: string | undefined,
) {
  // Clean-up the file
  await runtimeServiceDeleteFile(client, {
    path: newFilePath,
  });

  // Clean-up the `.env` file
  if (!originalEnvBlob) {
    // If .env file didn't exist before, delete it
    await runtimeServiceDeleteFile(client, {
      path: ".env",
    });
  } else {
    // If .env file existed before, restore its original content
    await runtimeServicePutFile(client, {
      path: ".env",
      blob: originalEnvBlob,
      create: true,
      createOnly: false,
    });
  }
}

async function setOlapConnectorInRillYAML(
  queryClient: QueryClient,
  client: RuntimeClient,
  newConnectorName: string,
): Promise<void> {
  await runtimeServicePutFile(client, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithOlapConnector(
      client,
      queryClient,
      newConnectorName,
    ),
    create: true,
    createOnly: false,
  });
}

async function setAiConnectorInRillYAML(
  queryClient: QueryClient,
  client: RuntimeClient,
  newConnectorName: string,
): Promise<void> {
  await runtimeServicePutFile(client, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithAiConnector(
      client,
      queryClient,
      newConnectorName,
    ),
    create: true,
    createOnly: false,
  });
}

async function saveConnectorWithoutTest(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  newConnectorName: string,
  client: RuntimeClient,
  existingEnvBlob?: string,
): Promise<void> {
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
  savedWithoutTestPaths.add(newConnectorFilePath);

  // Update .env file with secrets (keep ordering consistent with Test and Connect).
  // When existingEnvBlob is provided (e.g. Save overriding an in-flight Test and Connect),
  // use it as the baseline so env var names stay consistent and don't get _1 suffixes.
  const { newBlob: newEnvBlob, originalBlob: envBlobForYaml } =
    await updateDotEnvWithSecrets(client, queryClient, connector, formValues, {
      secretKeys: schemaSecretKeys,
      schema: schema ?? undefined,
      existingEnvBlob: existingEnvBlob,
    });

  await runtimeServicePutFile(client, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Always create/overwrite to ensure the connector file is created immediately
  await runtimeServicePutFile(client, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, formValues, {
      connectorInstanceName: newConnectorName,
      orderedProperties: schemaFields,
      secretKeys: schemaSecretKeys,
      stringKeys: schemaStringKeys,
      schema: schema ?? undefined,
      existingEnvBlob: envBlobForYaml,
      fieldFilter: schemaFields
        ? (property) => !("internal" in property && property.internal)
        : undefined,
    }),
    create: true,
    createOnly: false,
  });

  if (OLAP_ENGINES.includes(connector.name as string)) {
    await setOlapConnectorInRillYAML(queryClient, client, newConnectorName);
  }

  if (AI_CONNECTORS.includes(connector.name as string)) {
    await setAiConnectorInRillYAML(queryClient, client, newConnectorName);
  }

  // Go to the new connector file
  await goto(`/files/${newConnectorFilePath}`);
}

export async function submitAddConnectorForm(
  client: RuntimeClient,
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  saveAnyway: boolean = false,
  existingEnvBlob?: string,
): Promise<string> {
  await beforeSubmitForm(client, connector);
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
  const uniqueConnectorSubmissionKey = `${client.instanceId}:${connector.name}`;

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
      // If Save is clicked while Test and Connect is running,
      // proceed immediately without waiting for the ongoing operation
      connectorSubmissions.delete(uniqueConnectorSubmissionKey);

      // Use the same connector name from the ongoing operation
      const newConnectorName = existingSubmission.connectorName;

      // Proceed immediately with Save logic.
      // Pass existingEnvBlob so env var names stay consistent with what T&C used;
      // without it, re-reading .env (which T&C already modified) would generate
      // duplicate suffixed env var names.
      await saveConnectorWithoutTest(
        queryClient,
        connector,
        formValues,
        newConnectorName,
        client,
        existingEnvBlob,
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
      // Use originalBlob from updateDotEnvWithSecrets for consistent conflict detection
      const envResult = await updateDotEnvWithSecrets(
        client,
        queryClient,
        connector,
        formValues,
        {
          secretKeys: schemaSecretKeys,
          schema: schema ?? undefined,
        },
      );
      const newEnvBlob = envResult.newBlob;
      originalEnvBlob = envResult.originalBlob;

      if (saveAnyway) {
        // Save: bypass reconciliation entirely via centralized helper
        await saveConnectorWithoutTest(
          queryClient,
          connector,
          formValues,
          newConnectorName,
          client,
          existingEnvBlob,
        );
        return newConnectorName;
      }

      /**
       * Optimistic updates (Test and Connect):
       * 1. Write the `.env` file so secrets exist before connector reconciliation
       * 2. Create the `<connector>.yaml` file
       * 3. Wait for reconciliation and surface any file errors
       */
      await runtimeServicePutFileAndWaitForReconciliation(client, {
        path: ".env",
        blob: newEnvBlob,
        create: true,
        createOnly: false,
      });
      envWritten = true;

      await runtimeServicePutFile(client, {
        path: newConnectorFilePath,
        blob: compileConnectorYAML(connector, formValues, {
          connectorInstanceName: newConnectorName,
          orderedProperties: schemaFields,
          secretKeys: schemaSecretKeys,
          stringKeys: schemaStringKeys,
          schema: schema ?? undefined,
          existingEnvBlob: originalEnvBlob,
          fieldFilter: schemaFields
            ? (property) => !("internal" in property && property.internal)
            : undefined,
        }),
        create: true,
        createOnly: false,
      });
      connectorCreated = true;

      // Wait for connector resource-level reconciliation
      // This must happen after .env reconciliation since connectors depend on secrets
      await waitForResourceReconciliation(
        client,
        newConnectorName,
        ResourceKind.Connector,
      );

      // Check for file errors
      // If the connector file has errors, rollback the changes
      const errorMessage = await fileArtifacts.checkFileErrors(
        queryClient,
        newConnectorFilePath,
      );
      if (errorMessage) {
        throw new Error(errorMessage);
      }

      if (OLAP_ENGINES.includes(connector.name as string)) {
        await setOlapConnectorInRillYAML(queryClient, client, newConnectorName);
      }

      if (AI_CONNECTORS.includes(connector.name as string)) {
        await setAiConnectorInRillYAML(queryClient, client, newConnectorName);
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
        !savedWithoutTestPaths.has(newConnectorFilePath) &&
        (envWritten || connectorCreated);

      if (shouldRollbackConnectorFile) {
        await rollbackChanges(client, newConnectorFilePath, originalEnvBlob);
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
      // so a subsequent Save can still reuse the same connector file
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

/**
 * Update an existing connector: write new YAML and .env values,
 * then optionally test the connection via reconciliation.
 *
 * When saveOnly is true, the connector file is written without waiting
 * for reconciliation (same as the "Save" button for new connectors).
 */
export async function submitEditConnectorForm(
  client: RuntimeClient,
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  connectorInstanceName: string,
  saveOnly: boolean = false,
  existingEnvBlob?: string,
): Promise<void> {
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

  const connectorFilePath = getFileAPIPathFromNameAndType(
    connectorInstanceName,
    EntityType.Connector,
  );

  // Fetch the existing resource to extract env var names from YAML template
  // references (e.g. dsn: "{{ .env.POSTGRES_DSN_1 }}"). This preserves the
  // original env var names on re-save, avoiding conflicts with other connectors.
  const resource = await runtimeServiceGetResource(client, {
    name: { kind: ResourceKind.Connector, name: connectorInstanceName },
  });
  const specProperties = resource?.resource?.connector?.spec?.properties ?? {};

  // Build a mapping of property key → existing env var name from the YAML
  const existingEnvVarNames = new Map<string, string>();
  for (const key of schemaSecretKeys) {
    const templateRef = specProperties[key];
    if (typeof templateRef === "string") {
      const match = templateRef.match(/\{\{\s*\.env\.(\w+)\s*\}\}/);
      if (match?.[1]) {
        existingEnvVarNames.set(key, match[1]);
      }
    }
  }

  // Read the current .env
  let envBlob: string;
  if (existingEnvBlob !== undefined) {
    envBlob = existingEnvBlob;
  } else {
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGetFileQueryKey(client.instanceId, {
        path: ".env",
      }),
    });
    try {
      const file = await queryClient.fetchQuery({
        queryKey: getRuntimeServiceGetFileQueryKey(client.instanceId, {
          path: ".env",
        }),
        queryFn: () => runtimeServiceGetFile(client, { path: ".env" }),
      });
      envBlob = file.blob || "";
    } catch {
      envBlob = "";
    }
  }

  // Update .env: reuse existing env var names for secrets that already have
  // a template reference; use base names for new secrets.
  let newEnvBlob = envBlob;
  // Build a fake env blob for compileConnectorYAML that maps base names to
  // existing names. We achieve this by passing undefined (no blob) and then
  // doing a post-hoc replacement of env var references in the YAML.
  const envVarReplacements = new Map<string, string>();
  for (const key of schemaSecretKeys) {
    const value = formValues[key];
    if (!value) continue;
    const existingName = existingEnvVarNames.get(key);
    const baseName = getGenericEnvVarName(
      connector.name as string,
      key,
      schema ?? undefined,
    );
    const envVarName = existingName ?? baseName;
    newEnvBlob = replaceOrAddEnvVariable(
      newEnvBlob,
      envVarName,
      value as string,
    );
    // Track if base name differs from existing so we can fix the YAML
    if (existingName && existingName !== baseName) {
      envVarReplacements.set(baseName, existingName);
    }
  }

  // Generate YAML with base env var names (existingEnvBlob=undefined skips
  // conflict detection), then replace with actual names where they differ.
  let connectorYAML = compileConnectorYAML(connector, formValues, {
    connectorInstanceName,
    orderedProperties: schemaFields,
    secretKeys: schemaSecretKeys,
    stringKeys: schemaStringKeys,
    schema: schema ?? undefined,
    existingEnvBlob: undefined,
    fieldFilter: schemaFields
      ? (property) => !("internal" in property && property.internal)
      : undefined,
  });
  for (const [baseName, actualName] of envVarReplacements) {
    connectorYAML = connectorYAML.replace(
      `{{ .env.${baseName} }}`,
      `{{ .env.${actualName} }}`,
    );
  }

  if (saveOnly) {
    // Write .env and connector YAML without testing
    await runtimeServicePutFile(client, {
      path: ".env",
      blob: newEnvBlob,
      create: true,
      createOnly: false,
    });

    await runtimeServicePutFile(client, {
      path: connectorFilePath,
      blob: connectorYAML,
      create: true,
      createOnly: false,
    });
    return;
  }

  // Write .env first (secrets must exist before connector reconciliation)
  await runtimeServicePutFileAndWaitForReconciliation(client, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Write updated connector YAML
  await runtimeServicePutFile(client, {
    path: connectorFilePath,
    blob: connectorYAML,
    create: true,
    createOnly: false,
  });

  // Wait for reconciliation to test the connection
  await waitForResourceReconciliation(
    client,
    connectorInstanceName,
    ResourceKind.Connector,
  );

  // Check for file errors after reconciliation
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    connectorFilePath,
  );
  if (errorMessage) {
    throw new Error(errorMessage);
  }
}

export async function submitAddSourceForm(
  client: RuntimeClient,
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: AddDataFormValues,
  connectorInstanceName?: string,
): Promise<void> {
  await beforeSubmitForm(client, connector);
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

  // Create model YAML file
  const newSourceFilePath = getFileAPIPathFromNameAndType(
    newSourceName,
    EntityType.Table,
  );
  sourceIngestionTracker.trackPending(`/${newSourceFilePath}`);
  await runtimeServicePutFile(client, {
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

  // Create or update the `.env` file
  const { newBlob: newEnvBlob, originalBlob: originalEnvBlob } =
    await updateDotEnvWithSecrets(
      client,
      queryClient,
      rewrittenConnector,
      rewrittenFormValues,
      {
        secretKeys: schemaSecretKeys,
      },
    );

  // Make sure the file has reconciled before testing the connection
  await runtimeServicePutFileAndWaitForReconciliation(client, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Wait for source resource-level reconciliation
  // This must happen after .env reconciliation since sources depend on secrets
  try {
    await waitForResourceReconciliation(
      client,
      newSourceName,
      ResourceKind.Model,
    );
  } catch (error) {
    // The source file was already created, so we need to delete it
    sourceIngestionTracker.trackCancelled(`/${newSourceFilePath}`);
    await rollbackChanges(client, newSourceFilePath, originalEnvBlob);
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
    newSourceFilePath,
  );
  if (errorMessage) {
    sourceIngestionTracker.trackCancelled(`/${newSourceFilePath}`);
    await rollbackChanges(client, newSourceFilePath, originalEnvBlob);
    throw new Error(errorMessage);
  }

  await goto(`/files/${newSourceFilePath}`);
}
