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
import {
  getConnectorSchema,
  templateNameMap,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import { runtimeServiceGenerateFile } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

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
  connectorName,
  connector,
  schema,
  formValues,
  existingEnvBlob,
}: {
  connectorName: string;
  connector: V1ConnectorDriver;
  schema: MultiStepFormSchema | null;
  formValues: Record<string, unknown>;
  existingEnvBlob: string | null;
}) {
  const isPublicAuth = formValues.auth_method === "public";
  const [rewrittenConnector, rewrittenFormValues] = prepareSourceFormData(
    connector,
    formValues,
    {
      connectorInstanceName: isPublicAuth ? undefined : connectorName,
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

/**
 * Generate a YAML preview using the GenerateFile RPC for template-based connectors.
 * Returns the rendered model YAML from the backend template.
 */
export async function getTemplateYamlPreview(
  client: RuntimeClient,
  driverName: string,
  formValues: Record<string, unknown>,
  connectorName?: string,
): Promise<string> {
  const templateName = templateNameMap.get(driverName);
  if (!templateName) return "";

  try {
    const resp = await runtimeServiceGenerateFile(client, {
      templateName,
      output: "model",
      properties: formValues as Record<string, unknown>,
      connectorName: connectorName ?? "",
      preview: true,
    });
    const modelFile = resp.files?.find((f) => f.blob);
    return modelFile?.blob ?? "";
  } catch {
    return "";
  }
}
