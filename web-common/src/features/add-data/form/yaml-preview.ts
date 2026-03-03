import {
  filterSchemaValuesForSubmit,
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "@rilldata/web-common/features/templates/schema-utils.ts";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { compileConnectorYAML } from "@rilldata/web-common/features/connectors/code-utils.ts";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

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
