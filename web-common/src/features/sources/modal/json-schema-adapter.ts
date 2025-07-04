import type { Property, TemplateAPIResponse } from "./template-loader";

/**
 * Converts a template property to a JSON schema property
 */
function propertyToJsonSchema(property: Property): any {
  const baseSchema: any = {
    type: property.type === "number" ? "number" : property.type,
    title: property.displayName,
    description: property.hint,
  };

  // Add validation patterns if present
  if (property.validation?.pattern) {
    baseSchema.pattern = property.validation.pattern;
  }

  // Add required validation
  if (property.required) {
    baseSchema.minLength = 1;
  }

  // Add specific validations for different types
  if (property.type === "number") {
    if (property.validation?.pattern) {
      // Convert regex pattern to number validation
      baseSchema.pattern = undefined; // Remove pattern for numbers
      baseSchema.multipleOf = 1; // Ensure it's an integer if needed
    }
  }

  return baseSchema;
}

/**
 * Generates a JSON schema from template data for SuperForms validation
 */
export function generateJsonSchema(template: TemplateAPIResponse): any {
  const properties: Record<string, any> = {};
  const required: string[] = [];

  // Convert properties to JSON schema
  for (const property of template.properties) {
    properties[property.key] = propertyToJsonSchema(property);

    if (property.required) {
      required.push(property.key);
    }
  }

  const schema = {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    properties,
    required,
    additionalProperties: false,
  };

  return schema;
}

/**
 * Generates a JSON schema for DSN form
 */
export function generateDsnJsonSchema(template: TemplateAPIResponse): any {
  if (!template.dsn) {
    throw new Error("DSN template not available");
  }

  const schema = {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    properties: {
      [template.dsn.key]: propertyToJsonSchema(template.dsn),
    },
    required: template.dsn.required ? [template.dsn.key] : [],
    additionalProperties: false,
  };

  return schema;
}

/**
 * Creates a SuperForms validation adapter that conforms to the expected interface
 */
export function createJsonSchemaAdapter(schema: any) {
  return {
    superFormValidationLibrary: "json-schema",
    validate: (data: any) => {
      // This is a simplified validation - in production you'd use a proper JSON schema validator
      const errors: Record<string, string[]> = {};

      // Check required fields
      if (schema.required) {
        for (const field of schema.required) {
          if (
            !data[field] ||
            (typeof data[field] === "string" && data[field].trim() === "")
          ) {
            errors[field] = [
              `${schema.properties[field]?.title || field} is required`,
            ];
          }
        }
      }

      // Check patterns
      for (const [field, value] of Object.entries(data)) {
        const property = schema.properties[field];
        if (property?.pattern && typeof value === "string") {
          const regex = new RegExp(property.pattern);
          if (!regex.test(value)) {
            errors[field] = [
              property.description || `Invalid format for ${property.title}`,
            ];
          }
        }
      }

      return {
        valid: Object.keys(errors).length === 0,
        errors,
      };
    },
    schema,
  };
}

/**
 * Creates initial form data based on template properties
 */
export function createInitialFormData(
  template: TemplateAPIResponse,
): Record<string, any> {
  const initialData: Record<string, any> = {};

  for (const property of template.properties) {
    if (property.type === "boolean") {
      initialData[property.key] = false;
    } else {
      initialData[property.key] = "";
    }
  }

  return initialData;
}

/**
 * Creates initial DSN form data
 */
export function createInitialDsnFormData(
  template: TemplateAPIResponse,
): Record<string, any> {
  if (!template.dsn) {
    return {};
  }

  return {
    [template.dsn.key]: "",
  };
}
