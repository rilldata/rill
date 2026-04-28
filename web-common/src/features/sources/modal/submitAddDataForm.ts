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
  getRuntimeServiceGetInstanceQueryKey,
  runtimeServiceDeleteFile,
  runtimeServiceGetInstance,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import type { RuntimeClient } from "../../../runtime-client/v2";
import {
  compileConnectorYAML,
  replaceOrAddEnvVariable,
  updateDotEnvWithSecrets,
  updateRillYAMLWithAiConnector,
  updateRillYAMLWithOlapConnector,
} from "../../connectors/code-utils";
import {
  applyDuckLakeFormPipeline,
  injectDuckLakeAttach,
} from "../../templates/schemas/ducklake-utils";
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
  filterSchemaValuesForSubmit,
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
  schemaName?: string,
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
    // Determine the OLAP engine based on the connector being added. Prefer the
    // schema name over driver name so connectors that piggy-back on another
    // driver (e.g. DuckLake uses the duckdb driver) seed rill.yaml with their
    // own identity rather than the underlying driver.
    const effectiveSchemaName = schemaName ?? connector?.name ?? "";
    let olapEngine = "duckdb"; // Default for data sources

    if (connector && OLAP_ENGINES.includes(effectiveSchemaName)) {
      olapEngine = effectiveSchemaName;
    }

    await runtimeServiceUnpackEmpty(client, {
      displayName: EMPTY_PROJECT_TITLE,
      olap: olapEngine, // Explicitly set OLAP based on connector type
    });

    // Race condition: invalidate("app:init") must be called before we navigate to
    // `/files/${newFilePath}`. invalidate("app:init") is also called in the
    // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
    await invalidate("app:init");
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
  schemaName?: string,
): Promise<void> {
  const effectiveSchemaName = schemaName ?? connector.name ?? "";
  const schema = getConnectorSchema(effectiveSchemaName);
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
  const envResult = await updateDotEnvWithSecrets(
    client,
    queryClient,
    connector,
    formValues,
    {
      secretKeys: schemaSecretKeys,
      schema: schema ?? undefined,
      existingEnvBlob: existingEnvBlob,
    },
  );
  let newEnvBlob = envResult.newBlob;
  const envBlobForYaml = envResult.originalBlob;

  // DuckLake: compose Parameters tab into `attach`, route password fields and
  // raw-ATTACH catalog URIs through `.env`.
  const { transformedValues, extractedSecrets } = applyDuckLakeFormPipeline(
    schema,
    formValues,
    {
      connectorName: connector.name ?? "",
      existingEnvBlob: envBlobForYaml ?? "",
    },
  );
  for (const [envVarName, rawValue] of Object.entries(extractedSecrets)) {
    newEnvBlob = replaceOrAddEnvVariable(newEnvBlob, envVarName, rawValue);
  }

  // Re-inject `attach` after filtering — the tab-group filter drops it when
  // the "parameters" tab is active because `attach` belongs to the "sql" tab.
  const filteredValues = schema
    ? injectDuckLakeAttach(
        schema,
        filterSchemaValuesForSubmit(schema, transformedValues, {
          step: "connector",
        }),
        transformedValues,
      )
    : transformedValues;

  await runtimeServicePutFile(client, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Always create/overwrite to ensure the connector file is created immediately
  await runtimeServicePutFile(client, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, filteredValues, {
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

  if (OLAP_ENGINES.includes(effectiveSchemaName)) {
    await setOlapConnectorInRillYAML(queryClient, client, newConnectorName);
  }

  if (AI_CONNECTORS.includes(effectiveSchemaName)) {
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
  schemaName?: string,
): Promise<string> {
  // Prefer schemaName over connector.name (driver) so connectors that override
  // the backend driver (e.g. DuckLake uses the duckdb driver) resolve their
  // own schema, name their connector file after themselves, and route through
  // OLAP/AI bookkeeping under the right key.
  const effectiveSchemaName = schemaName ?? connector.name ?? "";
  await beforeSubmitForm(client, connector, effectiveSchemaName);
  const schema = getConnectorSchema(effectiveSchemaName);
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
  const uniqueConnectorSubmissionKey = `${client.instanceId}:${effectiveSchemaName}`;

  const newConnectorName = getName(
    effectiveSchemaName,
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
        effectiveSchemaName,
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
      let newEnvBlob = envResult.newBlob;
      originalEnvBlob = envResult.originalBlob;

      // DuckLake: compose Parameters tab into `attach`, route password fields
      // and raw-ATTACH catalog URIs through `.env` using the baseline blob so
      // env-var naming stays consistent with the form-field secrets that
      // updateDotEnvWithSecrets just processed.
      const duckLakeResult = applyDuckLakeFormPipeline(schema, formValues, {
        connectorName: connector.name ?? "",
        existingEnvBlob: originalEnvBlob ?? "",
      });
      formValues = duckLakeResult.transformedValues;
      for (const [envVarName, rawValue] of Object.entries(
        duckLakeResult.extractedSecrets,
      )) {
        newEnvBlob = replaceOrAddEnvVariable(newEnvBlob, envVarName, rawValue);
      }

      if (saveAnyway) {
        // Save: bypass reconciliation entirely via centralized helper
        await saveConnectorWithoutTest(
          queryClient,
          connector,
          formValues,
          newConnectorName,
          client,
          existingEnvBlob,
          effectiveSchemaName,
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

      // Re-inject `attach` after the tab-group filter, which drops it when
      // the "parameters" tab is active (since `attach` belongs to "sql").
      const filteredValues = schema
        ? injectDuckLakeAttach(
            schema,
            filterSchemaValuesForSubmit(schema, formValues, {
              step: "connector",
            }),
            formValues,
          )
        : formValues;

      await runtimeServicePutFile(client, {
        path: newConnectorFilePath,
        blob: compileConnectorYAML(connector, filteredValues, {
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

      if (OLAP_ENGINES.includes(effectiveSchemaName)) {
        await setOlapConnectorInRillYAML(queryClient, client, newConnectorName);
      }

      if (AI_CONNECTORS.includes(effectiveSchemaName)) {
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

      const errorDetails = error.details;
      if (errorDetails && errorDetails !== error.message) {
        throw {
          message: error.message || "Unable to establish a connection",
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

  // Get the default OLAP connector for the output block
  const runtimeInstance = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetInstanceQueryKey(client.instanceId, {}),
    queryFn: () => runtimeServiceGetInstance(client, { sensitive: false }),
  });
  const defaultOLAP = runtimeInstance?.instance?.olapConnector || "duckdb";

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
      outputConnector: defaultOLAP,
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
    const errorDetails = error.details;

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
