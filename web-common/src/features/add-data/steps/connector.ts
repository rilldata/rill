import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
  type V1ConnectorDriver,
} from "@rilldata/web-common/runtime-client";
import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized.ts";
import {
  runtimeServicePutFileAndWaitForReconciliation,
  waitForProjectParser,
  waitForResourceReconciliation,
} from "@rilldata/web-common/features/entity-management/actions.ts";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
import { OLAP_ENGINES } from "@rilldata/web-common/features/sources/modal/constants.ts";
import { invalidate } from "$app/navigation";
import {
  getConnectorSchema,
  isMultiStepConnector,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import {
  findRadioEnumKey,
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "@rilldata/web-common/features/templates/schema-utils.ts";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.ts";
import {
  compileConnectorYAML,
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "@rilldata/web-common/features/connectors/code-utils.ts";
import type { QueryClient } from "@tanstack/svelte-query";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export async function createConnector({
  runtimeClient,
  queryClient,
  connectorName,
  connectorDriver,
  formValues,
  validate,
  existingEnvBlob,
}: {
  runtimeClient: RuntimeClient;
  queryClient: QueryClient;
  connectorName: string;
  connectorDriver: V1ConnectorDriver;
  formValues: Record<string, unknown>;
  validate: boolean;
  existingEnvBlob: string | null;
}) {
  await maybeInitProject(runtimeClient, connectorDriver);

  const schema = getConnectorSchema(connectorDriver.name ?? "");

  // Fast-path: public auth skips validation/test and advances directly
  if (isMultiStepConnector(schema) && isPublicAuth(schema, formValues)) {
    return connectorDriver.name!;
  }

  const schemaFields = schema
    ? getSchemaFieldMetaList(schema, { step: "connector" })
    : [];
  const schemaSecretKeys = schema
    ? getSchemaSecretKeys(schema, { step: "connector" })
    : [];
  const schemaStringKeys = schema
    ? getSchemaStringKeys(schema, { step: "connector" })
    : [];

  // Create connector file path outside try block for cleanup
  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    connectorName,
    EntityType.Connector,
  );

  try {
    // Capture original .env and compute updated contents up front
    // Use originalBlob from updateDotEnvWithSecrets for consistent conflict detection
    const { newBlob: newEnvBlob, originalBlob: originalEnvBlob } =
      await updateDotEnvWithSecrets(
        runtimeClient,
        queryClient,
        connectorDriver,
        formValues,
        {
          secretKeys: schemaSecretKeys,
          schema: schema ?? undefined,
          existingEnvBlob: existingEnvBlob ?? undefined,
        },
      );

    /**
     * Optimistic updates (Test and Connect):
     * 1. Write the `.env` file so secrets exist before connector reconciliation
     * 2. Create the `<connector>.yaml` file
     * 3. Wait for reconciliation and surface any file errors
     */
    await runtimeServicePutFileAndWaitForReconciliation(runtimeClient, {
      path: ".env",
      blob: newEnvBlob,
      create: true,
      createOnly: false,
    });

    await runtimeServicePutFile(runtimeClient, {
      path: newConnectorFilePath,
      blob: compileConnectorYAML(connectorDriver, formValues, {
        connectorInstanceName: connectorName,
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

    if (validate) {
      // Wait for connector resource-level reconciliation
      // This must happen after .env reconciliation since connectors depend on secrets
      await waitForResourceReconciliation(
        runtimeClient,
        connectorName,
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
    }

    if (OLAP_ENGINES.includes(connectorDriver.name as string)) {
      await setOlapConnectorInRillYAML(
        queryClient,
        runtimeClient,
        connectorName,
      );
    }

    return newConnectorFilePath;
  } catch (error) {
    const errorDetails = error.details;
    if (errorDetails && errorDetails !== error.message) {
      throw {
        message: error.message || "Unable to establish a connection",
        details: errorDetails,
      };
    }

    throw error;
  }
}

async function maybeInitProject(
  client: RuntimeClient,
  connector: V1ConnectorDriver,
) {
  // If project is uninitialized, initialize an empty project
  const projectInitialized = await isProjectInitialized(client);
  if (projectInitialized) return;
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
  await waitForProjectParser(client.instanceId);

  // Race condition: invalidate("init") must be called before we navigate to
  // `/files/${newFilePath}`. invalidate("init") is also called in the
  // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
  await invalidate("init");
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

function isPublicAuth(
  schema: MultiStepFormSchema | null,
  values: Record<string, unknown>,
) {
  // Resolve the auth method from form values or the parent component's state
  const authKey = schema ? findRadioEnumKey(schema) : null;
  const selectedAuthMethod =
    (authKey && values?.[authKey] != null
      ? String(values[authKey])
      : undefined) || "";
  return selectedAuthMethod === "public";
}
