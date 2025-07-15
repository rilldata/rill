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
import { testOLAPConnector } from "../../connectors/olap/test-connection";
import { runtimeServicePutFileAndWaitForReconciliation } from "../../entity-management/actions";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { getName } from "../../entity-management/name-utils";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitialized } from "../../welcome/is-project-initialized";
import { compileSourceYAML, maybeRewriteToDuckDb } from "../sourceUtils";
import {
  testListTables,
  testGetTable,
  testListDatabaseSchemas,
} from "../../connectors/olap/test-connection";

interface AddDataFormValues {
  // name: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: unknown;
}

// Decide which test API(s) to call based on connector type and form values
export async function testSourceConnection(
  instanceId: string,
  connector: V1ConnectorDriver,
  formValues: Record<string, any>,
) {
  if (!connector.name) {
    throw new Error(
      "Connector name is required for testing source connection.",
    );
  }

  // FIXME
  const fileBased = ["s3", "gcs", "azure", "https", "local_file"];
  // FIXME
  const dbBased = [
    "sqlite",
    "duckdb",
    "bigquery",
    "athena",
    "redshift",
    "postgres",
    "mysql",
    "snowflake",
  ];

  if (fileBased.includes(connector.name)) {
    if (formValues.path) {
      return await testGetTable(instanceId, {
        connector: connector.name,
        table: String(formValues.path),
      });
    }
    return await testListTables(instanceId, {
      connector: connector.name,
    });
  }

  if (dbBased.includes(connector.name)) {
    if (formValues.schema) {
      const schemaResult = await testListDatabaseSchemas(instanceId, {
        connector: connector.name,
      });
      if (!schemaResult.success) return schemaResult;
    }
    const tablesParams: any = { connector: connector.name };
    if (formValues.schema) tablesParams.schema = String(formValues.schema);
    const tablesResult = await testListTables(instanceId, tablesParams);
    if (!tablesResult.success) return tablesResult;
    if (formValues.table) {
      const getTableParams: any = { connector: connector.name };
      if (formValues.schema) getTableParams.schema = String(formValues.schema);
      getTableParams.table = String(formValues.table);
      return await testGetTable(instanceId, getTableParams);
    }
    return tablesResult;
  }

  // Fallback: just try listing tables
  return await testListTables(instanceId, {
    connector: connector.name,
  });
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

  // Create or update the `.env` file and wait for reconciliation
  await runtimeServicePutFileAndWaitForReconciliation(instanceId, {
    path: ".env",
    blob: await updateDotEnvWithSecrets(
      queryClient,
      rewrittenConnector,
      rewrittenFormValues,
      "source",
    ),
    create: true,
    createOnly: false,
  });

  // Check for file errors
  const errorMessage = await fileArtifacts.checkFileErrors(
    queryClient,
    instanceId,
    newSourceFilePath,
  );
  if (errorMessage) {
    await rollbackConnectorChanges(instanceId, newSourceFilePath, undefined);
    throw new Error(errorMessage);
  }

  // Test the connection to the source using the strategy function
  const testResult = await testSourceConnection(
    instanceId,
    rewrittenConnector,
    rewrittenFormValues,
  );
  if (!testResult.success) {
    await rollbackConnectorChanges(instanceId, newSourceFilePath, undefined);
    throw {
      message: testResult.error || "Unable to establish a connection",
      details: testResult.details,
    };
  }

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
    blob: compileConnectorYAML(connector, formValues),
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
  // Make sure the file has reconciled before testing the connection
  const newEnvBlob = await updateDotEnvWithSecrets(
    queryClient,
    connector,
    formValues,
    "connector",
  );
  await runtimeServicePutFileAndWaitForReconciliation(instanceId, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  /**
   * Test the new OLAP connector:
   * 1. Ensure the file has reconciled and has no errors
   * 2. Test the connection to the OLAP database
   */

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

  // Test the connection to the OLAP database
  // If the connection test fails, rollback the changes
  const result = await testOLAPConnector(instanceId, {
    connector: connector.name as string,
  });
  if (!result.success) {
    await rollbackConnectorChanges(
      instanceId,
      newConnectorFilePath,
      originalEnvBlob,
    );
    throw {
      message: result.error || "Unable to establish a connection",
      details: result.details,
    };
  }

  /**
   * Connection successful: Complete the setup
   * Update the project configuration and navigate to the new connector
   */

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
