import {
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
import type { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";

export function getConnectorYamlPreview({
  connector,
  schema,
  formValues,
  envEditSession,
}: {
  connector: V1ConnectorDriver;
  schema: MultiStepFormSchema | null;
  formValues: Record<string, unknown>;
  envEditSession: EnvEditSession;
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
  const yamlPreview = compileConnectorYAML(
    connector,
    formValues,
    envEditSession,
    {
      fieldFilter: (property) => {
        if ("internal" in property && property.internal) return false;
        return !("noPrompt" in property && property.noPrompt);
      },
      orderedProperties: schemaFields,
      secretKeys: schemaSecretKeys,
      stringKeys: schemaStringKeys,
      schema: schema ?? undefined,
    },
  );

  return yamlPreview;
}

export function getSourceYamlPreview({
  connectorName,
  connector,
  schema,
  formValues,
  envEditSession,
  outputConnector,
}: {
  connectorName: string;
  connector: V1ConnectorDriver;
  schema: MultiStepFormSchema | null;
  formValues: Record<string, unknown>;
  envEditSession: EnvEditSession;
  outputConnector?: string;
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
    return compileSourceYAML(
      rewrittenConnector,
      rewrittenFormValues,
      envEditSession,
      {
        secretKeys: rewrittenSecretKeys,
        stringKeys: rewrittenStringKeys,
        originalDriverName: connector.name || undefined,
        outputConnector,
      },
    );
  }
  return getConnectorYamlPreview({
    connector,
    schema,
    formValues: rewrittenFormValues,
    envEditSession,
  });
}
