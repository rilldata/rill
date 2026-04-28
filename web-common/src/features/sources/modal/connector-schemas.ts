import type { V1ConnectorDriver } from "../../../runtime-client";
import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "../../templates/schemas/types";
import type { ConnectorStep } from "./connectorStepStore";
import { claudeSchema } from "../../templates/schemas/claude";
import { geminiSchema } from "../../templates/schemas/gemini";
import { openaiSchema } from "../../templates/schemas/openai";
import { ducklakeSchema } from "../../templates/schemas/ducklake";
import { connectorKeywordMapping } from "@rilldata/web-common/features/connectors/connector-metadata.ts";

/**
 * Connector schemas registered for synchronous lookup. Source-connector schemas
 * are populated dynamically via `registerTemplateSchema` when the `ListTemplates`
 * RPC resolves; AI connectors and DuckLake (which has client-side composer logic
 * in `ducklake-utils.ts`) stay as static imports because their flows are not
 * driven by the runtime templates registry.
 */
export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  ducklake: ducklakeSchema,
  claude: claudeSchema,
  openai: openaiSchema,
  gemini: geminiSchema,
};

/**
 * Connector information derived from JSON schemas.
 */
export interface ConnectorInfo {
  name: string;
  displayName: string;
  category: ConnectorCategory;
  keywords: string[];
}

/**
 * Map of connector names to ConnectorInfo objects, populated dynamically by
 * `registerTemplateSchema` when `ListTemplates` resolves. We need connector
 * info by name in a lot of places, so we have a map to optimize lookups.
 */
export const connectorInfoMap = new Map<string, ConnectorInfo>(
  Object.entries(multiStepFormSchemas)
    .filter(([, schema]) => schema?.["x-category"])
    .map(([name, schema]) => [
      name,
      {
        name,
        displayName: schema.title ?? name,
        category: schema["x-category"] as ConnectorCategory,
        keywords: connectorKeywordMapping[name] ?? [],
      },
    ]),
);

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema = multiStepFormSchemas[connectorName];
  return schema?.properties ? schema : null;
}

/**
 * Maps driver names to their full template names (e.g. "kafka" → "kafka-clickhouse").
 * Populated when templates are fetched from the ListTemplates RPC so that
 * AddDataFormManager can route to the right template for the active OLAP.
 */
export const templateNameMap = new Map<string, string>();

/**
 * Register a template schema dynamically. Called when templates are fetched
 * from the ListTemplates RPC so that connectors not in the static schema map
 * (e.g. kafka, hudi, mongodb when ClickHouse is the OLAP) work in the form
 * flow. Also updates connectorInfoMap so getConnectorDriverForSchema resolves.
 */
export function registerTemplateSchema(
  driverName: string,
  templateName: string,
  schema: MultiStepFormSchema,
  displayName: string,
) {
  multiStepFormSchemas[driverName] = schema;
  templateNameMap.set(driverName, templateName);
  const category = (schema["x-category"] ?? "sourceOnly") as ConnectorCategory;
  connectorInfoMap.set(driverName, {
    name: driverName,
    displayName,
    category,
    keywords: connectorKeywordMapping[driverName] ?? [],
  });
}

/**
 * Test seam: replace the schema cache with a fixture map. Used by specs that
 * need a deterministic set of schemas without invoking the runtime.
 */
export function populateSchemaCache(
  schemas: Record<string, MultiStepFormSchema>,
) {
  for (const [driverName, schema] of Object.entries(schemas)) {
    multiStepFormSchemas[driverName] = schema;
  }
}

/**
 * Get the backend driver name for a given schema name.
 * Returns x-driver if specified, otherwise returns the schema name.
 */
export function getBackendConnectorName(schemaName: string): string {
  const schema = getConnectorSchema(schemaName);
  return schema?.["x-driver"] ?? schemaName;
}

/**
 * Get the schema name for a given backend driver name.
 * Reverse lookup: finds the schema name that maps to the given driver name.
 * For most connectors, driver name = schema name, but some have x-driver override.
 */
export function getSchemaNameFromDriver(driverName: string): string {
  // First, check if driver name matches a schema name directly
  if (driverName in multiStepFormSchemas) {
    return driverName;
  }

  // If not, search for schema with matching x-driver
  for (const [schemaName, schema] of Object.entries(multiStepFormSchemas)) {
    const backendName = schema?.["x-driver"] ?? schemaName;
    if (backendName === driverName) {
      return schemaName;
    }
  }

  // Fallback: return driver name (assumes it's the schema name)
  return driverName;
}

/**
 * Determine if a connector has multi-step form flow (connector → source).
 * True for object store connectors (S3, GCS, Azure) and any schema that
 * defines fields on both the "connector" and "source" steps.
 */
export function isMultiStepConnector(
  schema: MultiStepFormSchema | null,
): boolean {
  if (!schema?.properties) return false;
  if (schema["x-category"] === "objectStore") return true;
  const fields = Object.values(schema.properties);
  return (
    fields.some((p) => p["x-step"] === "connector") &&
    fields.some((p) => p["x-step"] === "source")
  );
}

/**
 * Determine if a connector supports explorer mode (SQL query interface).
 * Detected by the presence of fields tagged with x-step: "explorer".
 */
export function hasExplorerStep(schema: MultiStepFormSchema | null): boolean {
  if (!schema?.properties) return false;
  return Object.values(schema.properties).some(
    (p) => p["x-step"] === "explorer",
  );
}

/**
 * Map a connector category to its docs URL path segment.
 */
export function getDocsCategory(
  category: ConnectorCategory | undefined,
): string {
  if (category === "ai") return "services";
  if (category === "olap") return "olap";
  return "data-source";
}

/**
 * Build a V1ConnectorDriver-compatible object from a schema name.
 */
export function toConnectorDriver(
  schemaName: string,
): V1ConnectorDriver | null {
  const schema = getConnectorSchema(schemaName);
  if (!schema) return null;
  const category = schema["x-category"];
  const backendName = getBackendConnectorName(schemaName);
  return {
    name: backendName,
    displayName: schema.title ?? schemaName,
    docsUrl: `https://docs.rilldata.com/developers/build/connectors/${getDocsCategory(category)}/${backendName}`,
    implementsObjectStore: category === "objectStore",
    implementsOlap: category === "olap",
    implementsSqlStore: category === "sqlStore",
    implementsWarehouse: category === "warehouse",
    implementsFileStore: category === "fileStore",
    implementsAi: category === "ai",
  };
}

/**
 * Get the form height CSS class for a connector's add data modal.
 * Some connectors with more fields use a taller form.
 */
export function getFormHeight(schema: MultiStepFormSchema | null): string {
  const FORM_HEIGHT_TALL = "max-h-[40rem] min-h-[40rem]";
  const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";
  return schema?.["x-form-height"] === "tall"
    ? FORM_HEIGHT_TALL
    : FORM_HEIGHT_DEFAULT;
}

/**
 * Connectors excluded from showing the "skip to import" link.
 * These connectors don't support skipping the connector setup step.
 */
export const SKIP_LINK_EXCLUDED_CONNECTORS = ["salesforce", "sqlite"];

/**
 * Determine if the skip link should be shown for a connector.
 * The skip link allows users to skip connector setup and go directly to import.
 * Only shown for connectors where handleSkip() can actually advance the step
 * (i.e., multi-step connectors or connectors with an explorer step).
 *
 * @param step - Current form step ("connector", "source", or "explorer")
 * @param connectorName - Name of the connector (e.g., "postgres", "s3")
 * @param connectorInstanceName - If set, user came from "Import Data" button
 * @param implementsOlap - Whether the connector is an OLAP engine
 */
export function shouldShowSkipLink(
  step: ConnectorStep,
  connectorName: string | undefined,
  connectorInstanceName: string | null,
  implementsOlap: boolean | undefined,
): boolean {
  if (
    step !== "connector" ||
    connectorInstanceName ||
    implementsOlap ||
    SKIP_LINK_EXCLUDED_CONNECTORS.includes(connectorName ?? "")
  ) {
    return false;
  }

  // Only show skip link if handleSkip() can actually advance the step
  const schema = getConnectorSchema(connectorName ?? "");
  return isMultiStepConnector(schema) || hasExplorerStep(schema);
}

/**
 * Get the form width CSS class for a connector's add data modal.
 * Some connectors with templates or more content use a wider form.
 */
export function getFormWidth(schema: MultiStepFormSchema | null): string {
  return schema?.["x-form-width"] === "wide" ? "max-w-5xl" : "max-w-4xl";
}
