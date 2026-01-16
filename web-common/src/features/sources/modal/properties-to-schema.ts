import type { ConnectorDriverProperty } from "@rilldata/web-common/runtime-client";
import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
import type {
  MultiStepFormSchema,
  JSONSchemaField,
} from "../../templates/schemas/types";

/**
 * Converts ConnectorDriverProperty[] to JSON Schema format.
 *
 * This is a translation layer at the API boundary - the ONLY place where
 * ConnectorDriverPropertyType should be used. Once converted, JSON Schema
 * types ("string" | "number" | "boolean" | "object") are the single source of truth.
 *
 * TODO: Once backend sends JSON Schema directly, this function can be removed.
 */
export function propertiesToSchema(
  properties: ConnectorDriverProperty[],
  step?: "connector" | "source",
): MultiStepFormSchema {
  const schemaProperties: Record<string, JSONSchemaField> = {};
  const required: string[] = [];

  for (const prop of properties) {
    if (!prop.key) continue;

    // Convert backend enum to JSON Schema type (the single source of truth)
    const field: JSONSchemaField = {
      type: convertBackendTypeToJsonSchemaType(prop.type),
      title: prop.displayName || prop.key,
      description: prop.description,
      "x-placeholder": prop.placeholder,
      "x-hint": prop.hint,
      "x-secret": prop.secret || false,
      "x-docs-url": prop.docsUrl,
    };

    // Convert default value based on JSON Schema type from the field itself
    if (
      prop.default !== undefined &&
      prop.default !== null &&
      prop.default !== ""
    ) {
      if (field.type === "number") {
        const num = Number(prop.default);
        if (!isNaN(num)) {
          field.default = num;
        }
      } else if (field.type === "boolean") {
        field.default = prop.default === "true";
      } else {
        field.default = prop.default;
      }
    }

    // Map backend-specific types to JSON Schema extensions
    // (These are the only places we need to check the backend enum)
    if (prop.type === ConnectorDriverPropertyType.TYPE_FILE) {
      field["x-display"] = "file";
      field.format = "file";
      field["x-accept"] = ".json";
    } else if (prop.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL) {
      field["x-informational"] = true;
    }

    // Set step if provided
    if (step) {
      field["x-step"] = step;
    }

    schemaProperties[prop.key] = field;

    // Add to required array if the property is required
    if (prop.required) {
      required.push(prop.key);
    }
  }

  return {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    properties: schemaProperties,
    required,
  };
}

/**
 * Translation layer: converts backend enum to JSON Schema type.
 * This is the ONLY place ConnectorDriverPropertyType should be referenced.
 *
 * Once backend sends JSON Schema directly, this function can be removed.
 */
function convertBackendTypeToJsonSchemaType(
  type?: ConnectorDriverPropertyType,
): "string" | "number" | "boolean" | "object" {
  switch (type) {
    case ConnectorDriverPropertyType.TYPE_NUMBER:
      return "number";
    case ConnectorDriverPropertyType.TYPE_BOOLEAN:
      return "boolean";
    case ConnectorDriverPropertyType.TYPE_STRING:
    case ConnectorDriverPropertyType.TYPE_FILE:
    case ConnectorDriverPropertyType.TYPE_INFORMATIONAL:
    case ConnectorDriverPropertyType.TYPE_UNSPECIFIED:
    default:
      return "string";
  }
}
