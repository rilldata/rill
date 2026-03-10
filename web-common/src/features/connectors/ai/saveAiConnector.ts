import { goto } from "$app/navigation";
import type { QueryClient } from "@tanstack/query-core";
import { runtimeServicePutFile } from "../../../runtime-client";
import type { RuntimeClient } from "../../../runtime-client/v2";
import {
  compileConnectorYAML,
  updateDotEnvWithSecrets,
  updateRillYAMLWithAiConnector,
} from "../code-utils";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { getName } from "../../entity-management/name-utils";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { EntityType } from "../../entity-management/types";
import { beforeSubmitForm } from "../../sources/modal/submitAddDataForm";
import {
  getConnectorSchema,
  toConnectorDriver,
} from "../../sources/modal/connector-schemas";
import {
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "../../templates/schema-utils";

async function setAiConnectorInRillYAML(
  queryClient: QueryClient,
  client: RuntimeClient,
  newConnectorName: string,
): Promise<void> {
  await runtimeServicePutFile(client, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithAiConnector(queryClient, newConnectorName),
    create: true,
    createOnly: false,
  });
}

/**
 * Save an AI connector directly: write .env, connector YAML, update rill.yaml,
 * and navigate to the new file. Used by AddAiConnectorDialog.
 */
export async function saveAiConnector(
  client: RuntimeClient,
  queryClient: QueryClient,
  schemaName: string,
  formValues: Record<string, string>,
): Promise<void> {
  const connector = toConnectorDriver(schemaName);
  if (!connector) throw new Error(`Unknown AI connector: ${schemaName}`);

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

  const newConnectorName = getName(
    connector.name as string,
    fileArtifacts.getNamesForKind(ResourceKind.Connector),
  );

  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    newConnectorName,
    EntityType.Connector,
  );

  // Write secrets to .env
  const { newBlob: newEnvBlob, originalBlob: envBlobForYaml } =
    await updateDotEnvWithSecrets(client, queryClient, connector, formValues, {
      secretKeys: schemaSecretKeys,
      schema: schema ?? undefined,
    });

  await runtimeServicePutFile(client, {
    path: ".env",
    blob: newEnvBlob,
    create: true,
    createOnly: false,
  });

  // Write connector YAML
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

  // Register as the project's AI connector
  await setAiConnectorInRillYAML(queryClient, client, newConnectorName);

  await goto(`/files/${newConnectorFilePath}`);
}
