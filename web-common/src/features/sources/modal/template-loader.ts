import connectorTemplates from "./connector-templates.json";

export type PropertyValidation = {
  pattern: string;
  patternMessage: string;
};

export type Property = {
  key: string;
  type: "string" | "number" | "boolean";
  displayName: string;
  required: boolean;
  hint: string;
  docsUrl: string;
  secret: boolean;
  placeholder?: string;
  validation?: PropertyValidation;
  description?: string;
};

export type TemplateAPIResponse = {
  properties: Property[];
  dsn?: Property;
};

/**
 * Loads template data for a connector
 * In the future, this could fetch from an API endpoint
 * @param connectorName - The name of the connector
 * @returns Promise<TemplateAPIResponse>
 */
export async function loadTemplateData(
  connectorName: string,
): Promise<TemplateAPIResponse> {
  // TODO: Replace with actual API call when backend is ready
  // const response = await fetch(`/api/connectors/${connectorName}/template`);
  // const data = await response.json();

  // For now, return the template from our JSON file
  const templates = connectorTemplates as Record<string, TemplateAPIResponse>;
  const template = templates[connectorName];

  if (!template) {
    throw new Error(`No template found for connector: ${connectorName}`);
  }

  return template;
}

/**
 * Validates template data against the JSON schema
 * @param data - The template data to validate
 * @returns boolean indicating if data is valid
 */
export function validateTemplateData(
  data: unknown,
): data is TemplateAPIResponse {
  // TODO: Implement proper JSON schema validation
  // For now, just do basic type checking
  if (!data || typeof data !== "object") return false;

  const templateData = data as any;

  if (!Array.isArray(templateData.properties)) return false;

  // Validate each property
  for (const property of templateData.properties) {
    if (!property.key || !property.type || !property.displayName) return false;
    if (typeof property.required !== "boolean") return false;
    if (!property.hint || !property.docsUrl) return false;
    if (typeof property.secret !== "boolean") return false;
  }

  // Validate DSN if present
  if (templateData.dsn) {
    const dsn = templateData.dsn;
    if (!dsn.key || !dsn.type || !dsn.displayName) return false;
    if (typeof dsn.required !== "boolean") return false;
    if (!dsn.hint || !dsn.docsUrl) return false;
    if (typeof dsn.secret !== "boolean") return false;
  }

  return true;
}

/**
 * Gets template data for a specific connector
 * @param connectorName - The name of the connector
 * @returns TemplateAPIResponse or null if not found
 */
export async function getConnectorTemplate(
  connectorName: string,
): Promise<TemplateAPIResponse | null> {
  try {
    const data = await loadTemplateData(connectorName);

    if (!validateTemplateData(data)) {
      console.error(`Invalid template data for connector: ${connectorName}`);
      return null;
    }

    return data;
  } catch (error) {
    console.error(
      `Failed to load template for connector ${connectorName}:`,
      error,
    );
    return null;
  }
}

/**
 * Gets a list of available connector templates
 * @returns Array of connector names that have templates
 */
export function getAvailableConnectorTemplates(): string[] {
  const templates = connectorTemplates as Record<string, TemplateAPIResponse>;
  return Object.keys(templates);
}

/**
 * Checks if a connector has a template available
 * @param connectorName - The name of the connector
 * @returns boolean indicating if template exists
 */
export function hasConnectorTemplate(connectorName: string): boolean {
  const templates = connectorTemplates as Record<string, TemplateAPIResponse>;
  return connectorName in templates;
}
