import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGetInstanceQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceDeleteFile,
  runtimeServiceGetFile,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
  type V1ConnectorDriver,
  type V1GetInstanceResponse,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized.ts";
import {
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
  getSchemaSecretKeys,
} from "@rilldata/web-common/features/templates/schema-utils.ts";
import { applyDuckLakeFormPipeline } from "@rilldata/web-common/features/templates/schemas/ducklake-utils.ts";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import {
  addLeadingSlash,
  getFileAPIPathFromNameAndType,
} from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.ts";
import {
  maybeUnsetOlapConnectorInYaml,
  replaceOrAddEnvVariable,
  unsetResourceEnvVars,
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "@rilldata/web-common/features/connectors/code-utils.ts";
import type { QueryClient } from "@tanstack/svelte-query";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { getConnectorYamlPreview } from "@rilldata/web-common/features/add-data/form/yaml-preview.ts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import {
  getProjectParserVersion,
  waitForProjectParserVersion,
} from "@rilldata/web-common/features/entity-management/project-parser.ts";

export async function createConnector({
  runtimeClient,
  queryClient,
  connectorName,
  connectorDriver,
  schemaName,
  formValues,
  validate,
  existingEnvBlob,
}: {
  runtimeClient: RuntimeClient;
  queryClient: QueryClient;
  connectorName: string;
  connectorDriver: V1ConnectorDriver;
  schemaName?: string;
  formValues: Record<string, unknown>;
  validate: boolean;
  existingEnvBlob: string | null;
}) {
  await maybeInitProject(runtimeClient);

  // Prefer schemaName for schema lookup so connectors that override the
  // backend driver (e.g. DuckLake uses the duckdb driver) still resolve
  // their own schema fields.
  const schema = getConnectorSchema(schemaName ?? connectorDriver.name ?? "");

  // DuckLake: compose Parameters tab into `attach` (with password fields
  // stored in `.env` and referenced via template) and route raw-ATTACH
  // catalog URIs through `.env`. Done before updateDotEnvWithSecrets so the
  // same baseline blob drives env-var name conflict detection for all
  // ducklake-derived secrets and any form-field secrets.
  const duckLakeResult = applyDuckLakeFormPipeline(schema, formValues, {
    connectorName: connectorDriver.name ?? "",
    existingEnvBlob: existingEnvBlob ?? "",
  });
  formValues = duckLakeResult.transformedValues;
  const duckLakeAttachSecrets = duckLakeResult.extractedSecrets;

  // Fast-path: public auth skips validation/test and advances directly
  if (isMultiStepConnector(schema) && isPublicAuth(schema, formValues)) {
    return connectorDriver.name!;
  }

  const schemaSecretKeys = schema
    ? getSchemaSecretKeys(schema, { step: "connector" })
    : [];

  // Create connector file path outside try block for cleanup
  const newConnectorFilePath = addLeadingSlash(
    getFileAPIPathFromNameAndType(connectorName, EntityType.Connector),
  );

  try {
    // Capture original .env and compute updated contents up front
    // Use originalBlob from updateDotEnvWithSecrets for consistent conflict detection
    const { newBlob: initialEnvBlob, originalBlob: originalEnvBlob } =
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
    let newEnvBlob = initialEnvBlob;

    // Persist DuckLake catalog URIs extracted from the raw ATTACH clause.
    // These are not tied to a form field, so updateDotEnvWithSecrets cannot
    // write them; append directly using the same env blob so write ordering
    // matches the rest of the secret handling.
    for (const [envVarName, rawValue] of Object.entries(
      duckLakeAttachSecrets,
    )) {
      newEnvBlob = replaceOrAddEnvVariable(newEnvBlob, envVarName, rawValue);
    }

    /**
     * Optimistic updates (Test and Connect):
     * 1. Write the `.env` file so secrets exist before connector reconciliation
     * 2. Create the `<connector>.yaml` file
     * 3. Wait for reconciliation and surface any file errors
     */

    // Get the project parser starting version
    const projectParserStartingVersion = getProjectParserVersion(
      runtimeClient.instanceId,
    );
    // Get the starting version of the connector resource
    const connectorStartingVersion = queryClient.getQueryData<{
      resource: V1Resource | undefined;
    }>(
      getRuntimeServiceGetResourceQueryKey(runtimeClient.instanceId, {
        name: { name: connectorName, kind: ResourceKind.Connector },
      }),
    )?.resource?.meta?.stateVersion;

    await runtimeServicePutFile(runtimeClient, {
      path: ".env",
      blob: newEnvBlob,
      create: true,
      createOnly: false,
    });

    await runtimeServicePutFile(runtimeClient, {
      path: newConnectorFilePath,
      blob: getConnectorYamlPreview({
        connector: connectorDriver,
        formValues,
        schema,
        existingEnvBlob: originalEnvBlob,
      }),
      create: true,
      createOnly: false,
    });

    if (validate) {
      // Wait for project parser to finish updating before checking for errors.
      await waitForProjectParserVersion(
        runtimeClient.instanceId,
        projectParserStartingVersion + 1,
      );

      await waitForResourceReconciliation(
        runtimeClient,
        connectorName,
        ResourceKind.Connector,
        connectorStartingVersion,
      );

      // Check for file errors
      // If the connector file has errors, rollback the changes
      const errorMessage = fileArtifacts.checkFileErrors(
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

export async function maybeDeleteConnector(
  runtimeClient: RuntimeClient,
  queryClient: QueryClient,
  connectorName: string,
) {
  const connectorFilePath = addLeadingSlash(
    getFileAPIPathFromNameAndType(connectorName, EntityType.Connector),
  );
  if (!fileArtifacts.hasFileArtifact(connectorFilePath)) return;

  const connectorFileArtifact =
    fileArtifacts.getFileArtifact(connectorFilePath);
  const connectorYaml = await connectorFileArtifact.fetchContent();
  if (!connectorYaml) return;

  // Get the existing env and remove the connector's env vars
  const envBlob = await unsetResourceEnvVars(
    runtimeClient,
    queryClient,
    connectorYaml,
  );

  // Delete the connector file
  await runtimeServiceDeleteFile(runtimeClient, {
    path: connectorFilePath,
  });

  // Update the .env file with the removed env vars
  await runtimeServicePutFile(runtimeClient, {
    path: ".env",
    blob: envBlob,
  });

  // Update the rill.yaml file to remove the connector as the OLAP connector.
  await unsetOlapConnectorInRillYAML(runtimeClient, queryClient, connectorName);
}

export async function maybeInitProject(client: RuntimeClient) {
  // If project is uninitialized, initialize an empty project
  const projectInitialized = await isProjectInitialized(client);
  if (projectInitialized) return;

  await runtimeServiceUnpackEmpty(client, {
    displayName: EMPTY_PROJECT_TITLE,
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

const ConnectorUnsetCheckMaxRetries = 5;
const ConnectorUnsetCheckIntervalConstant = 300;
const ConnectorUnsetCheckIntervalMultiplier = 300;

async function unsetOlapConnectorInRillYAML(
  runtimeClient: RuntimeClient,
  queryClient: QueryClient,
  connectorName: string,
) {
  // Get the existing rill.yaml file
  const file = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetFileQueryKey(runtimeClient.instanceId, {
      path: "rill.yaml",
    }),
    queryFn: () => runtimeServiceGetFile(runtimeClient, { path: "rill.yaml" }),
  });
  const blob = file.blob || "";

  const [ok, newBlob] = maybeUnsetOlapConnectorInYaml(blob, connectorName);
  if (!ok) return;

  await runtimeServicePutFile(runtimeClient, {
    path: "rill.yaml",
    blob: newBlob,
  });

  // Wait for rill.yaml to be updated
  let retryCount = 0;
  while (retryCount < ConnectorUnsetCheckMaxRetries) {
    try {
      const runtimeInstanceResp =
        queryClient.getQueryData<V1GetInstanceResponse>(
          getRuntimeServiceGetInstanceQueryKey(runtimeClient.instanceId, {
            sensitive: true,
          }),
        );

      if (
        !runtimeInstanceResp?.instance || // type safety
        runtimeInstanceResp.instance.olapConnector === connectorName
      ) {
        // Connector is not changed yet
        throw new Error("Connector not updated");
      }
      // Connector is removed from rill.yaml
      break;
    } catch {
      retryCount++;
      await new Promise((resolve) =>
        setTimeout(
          resolve,
          ConnectorUnsetCheckIntervalConstant +
            retryCount * ConnectorUnsetCheckIntervalMultiplier,
        ),
      );
    }
  }
}

export function isPublicAuth(
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

type CacheEntry = {
  name: string;
  formValues: Record<string, unknown>;
  existingEnvBlob: string;
};

export class ConnectorFormCache {
  private id = 0;

  private cache = new Map<string, CacheEntry>();

  public getNextId() {
    const id = ++this.id;
    return id.toString();
  }

  public async getOrCreate(schema: string, id: string): Promise<CacheEntry> {
    if (this.cache.has(id)) {
      return this.cache.get(id)!;
    }

    const name = getName(
      schema,
      fileArtifacts.getNamesForKind(ResourceKind.Connector),
    );

    const envFile = fileArtifacts.getFileArtifact(".env");
    const envBlob = (await envFile.fetchContent(false)) ?? "";

    const entry = {
      name,
      formValues: {},
      existingEnvBlob: envBlob,
    };
    this.cache.set(id, entry);
    return entry;
  }

  public updateFormValues(id: string, formValues: Record<string, unknown>) {
    const entry = this.cache.get(id);
    if (entry) {
      entry.formValues = formValues;
    }
  }

  public clear() {
    this.cache.clear();
    this.id = 0;
  }
}
export const connectorFormCache = new ConnectorFormCache();
