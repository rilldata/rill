import { runtimeServiceGetResource } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getConnectorSchema } from "./connector-schemas";
import {
  getSchemaSecretKeys,
  getSchemaFieldMetaList,
} from "../../templates/schema-utils";

/**
 * Parse .env content into a key-value map.
 */
function parseEnvBlob(blob: string): Record<string, string> {
  const result: Record<string, string> = {};
  for (const line of blob.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eqIndex = trimmed.indexOf("=");
    if (eqIndex === -1) continue;
    result[trimmed.slice(0, eqIndex)] = trimmed.slice(eqIndex + 1);
  }
  return result;
}

/**
 * Check if any of the given fields have a non-empty value in either
 * specProperties or formValues.
 */
function hasPopulatedField(
  fields: string[],
  specProperties: Record<string, unknown>,
  formValues: Record<string, unknown>,
  schema: Record<string, any>,
): boolean {
  return fields.some((f: string) => {
    // Check specProperties (may contain template refs like {{ .env.X }})
    if (
      specProperties[f] !== undefined &&
      specProperties[f] !== null &&
      specProperties[f] !== ""
    )
      return true;
    // Check formValues (actual resolved values)
    if (
      formValues[f] !== undefined &&
      formValues[f] !== null &&
      formValues[f] !== ""
    )
      return true;
    // For x-ui-only fields, check if they were already inferred
    const prop = schema.properties?.[f];
    if (prop?.["x-ui-only"] && formValues[f]) return true;
    return false;
  });
}

/**
 * Infer the correct values for UI-only selector fields (connection_mode,
 * deployment_type, auth_method, etc.) from which properties are populated.
 *
 * Handles both x-tab-group and x-grouped-fields. Runs in two passes so
 * that nested selectors (e.g. connection_mode inside deployment_type's
 * grouped fields) are resolved before the parent checks them.
 */
function inferUiOnlyTabValues(
  schema: Record<string, any>,
  specProperties: Record<string, unknown>,
  formValues: Record<string, unknown>,
): void {
  if (!schema?.properties) return;

  // Collect all x-ui-only fields that need inference, split by depth:
  // fields with x-tab-group are leaf-level (resolve first), fields with
  // x-grouped-fields may reference other ui-only fields (resolve second).
  const tabGroupFields: [string, Record<string, any>][] = [];
  const groupedFields: [string, Record<string, any>][] = [];

  for (const [key, fieldDef] of Object.entries(schema.properties)) {
    const field = fieldDef as Record<string, any>;
    if (!field["x-ui-only"]) continue;
    if (field["x-tab-group"]) tabGroupFields.push([key, field]);
    else if (field["x-grouped-fields"]) groupedFields.push([key, field]);
  }

  // Pass 1: resolve x-tab-group fields (e.g. connection_mode)
  for (const [key, field] of tabGroupFields) {
    const groups = field["x-tab-group"] as Record<string, string[]>;
    for (const [tabValue, tabFields] of Object.entries(groups)) {
      if (hasPopulatedField(tabFields, specProperties, formValues, schema)) {
        formValues[key] = tabValue;
        break;
      }
    }
  }

  // Pass 2: resolve x-grouped-fields (e.g. deployment_type) which may
  // reference already-inferred ui-only children.
  for (const [key, field] of groupedFields) {
    const groups = field["x-grouped-fields"] as Record<string, string[]>;
    for (const [groupValue, childFields] of Object.entries(groups)) {
      if (hasPopulatedField(childFields, specProperties, formValues, schema)) {
        formValues[key] = groupValue;
        break;
      }
    }
  }
}

/**
 * Extract the env var name from a Go template reference like
 * `{{ .env.POSTGRES_DSN }}`. Returns undefined if the value
 * is not a template reference.
 */
function extractEnvVarFromTemplate(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined;
  const match = value.match(/\{\{\s*\.env\.(\w+)\s*\}\}/);
  return match?.[1];
}

/**
 * Load existing connector values from the resource and .env file
 * to pre-fill the edit form.
 *
 * Non-secret properties come from the resource's spec.properties.
 * Secret values are resolved by parsing the Go template references
 * in spec.properties (e.g. `{{ .env.POSTGRES_DSN }}`) and looking
 * up the actual env var in .env. This ensures we always read the
 * correct key, even for suffixed names like POSTGRES_DSN_1.
 * UI-only tab selectors (connection_mode, auth_method) are inferred
 * from which properties are populated.
 */
export async function loadConnectorFormValues(
  client: RuntimeClient,
  connectorInstanceName: string,
  schemaName: string,
  envBlob: string,
): Promise<Record<string, unknown>> {
  const schema = getConnectorSchema(schemaName);
  if (!schema) return {};

  // Fetch the connector resource to get its properties
  const resource = await runtimeServiceGetResource(client, {
    name: { kind: ResourceKind.Connector, name: connectorInstanceName },
  });

  const specProperties = resource?.resource?.connector?.spec?.properties ?? {};
  const secretKeys = new Set(
    getSchemaSecretKeys(schema, { step: "connector" }),
  );
  const allFields = getSchemaFieldMetaList(schema, { step: "connector" });
  const envMap = parseEnvBlob(envBlob);

  const formValues: Record<string, unknown> = {};

  for (const field of allFields) {
    const key = field.key;

    if (secretKeys.has(key)) {
      // Secret field: parse the env var name from the YAML template reference
      // (e.g. dsn: "{{ .env.POSTGRES_DSN_1 }}" → look up POSTGRES_DSN_1)
      const envVarName = extractEnvVarFromTemplate(specProperties[key]);
      if (envVarName && envMap[envVarName] !== undefined) {
        formValues[key] = envMap[envVarName];
      }
    } else {
      // Non-secret field: use from resource spec.properties
      const value = specProperties[key];
      if (value !== undefined && value !== null) {
        formValues[key] = value;
      }
    }
  }

  // Infer UI-only tab selectors (connection_mode, auth_method) from
  // which properties are populated; e.g. if dsn is set, select the DSN tab
  inferUiOnlyTabValues(schema, specProperties, formValues);

  return formValues;
}
