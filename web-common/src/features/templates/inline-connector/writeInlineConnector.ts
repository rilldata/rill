import type { QueryClient } from "@tanstack/query-core";
import {
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  runtimeServicePutFile,
  type V1ConnectorDriver,
} from "../../../runtime-client";
import type { RuntimeClient } from "../../../runtime-client/v2";
import {
  compileConnectorYAML,
  updateDotEnvWithSecrets,
} from "../../connectors/code-utils";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { EntityType } from "../../entity-management/types";
import { getConnectorSchema } from "../../sources/modal/connector-schemas";
import {
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "../schema-utils";

export type InlineConnectorDriver = "s3" | "gcs" | "azure";

export async function writeInlineConnector(opts: {
  client: RuntimeClient;
  queryClient: QueryClient;
  driver: InlineConnectorDriver;
  values: Record<string, unknown>;
  connectorName: string;
}): Promise<void> {
  const { client, queryClient, driver, values, connectorName } = opts;

  const schema = getConnectorSchema(driver);
  if (!schema) {
    throw new Error(`No schema registered for driver "${driver}"`);
  }

  const connector: V1ConnectorDriver = { name: driver };
  const secretKeys = getSchemaSecretKeys(schema, { step: "connector" });
  const stringKeys = getSchemaStringKeys(schema, { step: "connector" });
  const schemaFields = getSchemaFieldMetaList(schema, { step: "connector" });

  const { newBlob, originalBlob } = await updateDotEnvWithSecrets(
    client,
    queryClient,
    connector,
    values,
    { secretKeys, schema },
  );

  await runtimeServicePutFile(client, {
    path: ".env",
    blob: newBlob,
    create: true,
    createOnly: false,
  });

  await runtimeServicePutFile(client, {
    path: getFileAPIPathFromNameAndType(connectorName, EntityType.Connector),
    blob: compileConnectorYAML(connector, values, {
      connectorInstanceName: connectorName,
      orderedProperties: schemaFields,
      secretKeys,
      stringKeys,
      schema,
      existingEnvBlob: originalBlob,
    }),
    create: true,
    createOnly: false,
  });

  await queryClient.invalidateQueries({
    queryKey: getRuntimeServiceAnalyzeConnectorsQueryKey(client.instanceId, {}),
  });
}
