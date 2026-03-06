import {
  filterSchemaValuesForSubmit,
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "@rilldata/web-common/features/templates/schema-utils.ts";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { compileConnectorYAML } from "@rilldata/web-common/features/connectors/code-utils.ts";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import {
  compileSourceYAML,
  prepareSourceFormData,
} from "@rilldata/web-common/features/sources/sourceUtils.ts";
import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";

export function getConnectorYamlPreview({
  connector,
  schema,
  formValues,
  existingEnvBlob,
}: {
  connector: V1ConnectorDriver;
  schema: MultiStepFormSchema | null;
  formValues: Record<string, unknown>;
  existingEnvBlob: string | null;
}) {
  const schemaFields = schema
    ? getSchemaFieldMetaList(schema, { step: "connector" })
    : [];
  const schemaSecretKeys = schema
    ? getSchemaSecretKeys(schema, { step: "connector" })
    : [];
  const schemaStringKeys = schema
    ? getSchemaStringKeys(schema, { step: "connector" })
    : [];
  const filteredValues = schema
    ? filterSchemaValuesForSubmit(schema, formValues, { step: "connector" })
    : formValues;
  const yamlPreview = compileConnectorYAML(connector, filteredValues, {
    fieldFilter: (property) => {
      if ("internal" in property && property.internal) return false;
      return !("noPrompt" in property && property.noPrompt);
    },
    orderedProperties: schemaFields,
    secretKeys: schemaSecretKeys,
    stringKeys: schemaStringKeys,
    schema: schema ?? undefined,
    existingEnvBlob: existingEnvBlob ?? "",
  });

  return yamlPreview;
}

export function getSourceYamlPreview({
  connector,
  schema,
  formValues,
  existingEnvBlob,
}: {
  connector: V1ConnectorDriver;
  schema: MultiStepFormSchema | null;
  formValues: Record<string, unknown>;
  existingEnvBlob: string | null;
}) {
  const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
    connector,
    formValues,
    {
      connectorInstanceName: connector.name,
    },
  );
  const isRewrittenToDuckDb = rewrittenConnector.name === "duckdb";
  const rewrittenSchema = getConnectorSchema(rewrittenConnector.name ?? "");
  const rewrittenSecretKeys = rewrittenSchema
    ? getSchemaSecretKeys(rewrittenSchema, { step: "source" })
    : undefined;
  const rewrittenStringKeys = rewrittenSchema
    ? getSchemaStringKeys(rewrittenSchema, { step: "source" })
    : undefined;
  if (isRewrittenToDuckDb) {
    return compileSourceYAML(rewrittenConnector, rewrittenFormValues, {
      secretKeys: rewrittenSecretKeys,
      stringKeys: rewrittenStringKeys,
      originalDriverName: connector.name || undefined,
    });
  }
  return getConnectorYamlPreview({
    connector,
    schema,
    formValues: rewrittenFormValues,
    existingEnvBlob,
  });
}
