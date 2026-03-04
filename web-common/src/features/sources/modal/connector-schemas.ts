import {
  createRuntimeServiceListTemplates,
  type V1Template,
} from "../../../runtime-client";
import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "../../templates/schemas/types";
import { SOURCES, OLAP_ENGINES } from "./constants";
import { derived, type Readable } from "svelte/store";

const SOURCES_SET = new Set(SOURCES);
const OLAP_SET = new Set(OLAP_ENGINES);

// Module-level cache populated when the TanStack Query resolves.
// Safe because AddDataModal (the entry point) subscribes to the query
// and renders the connector grid (step 1) first; by the time step 2
// needs getConnectorSchema(), the cache is populated.
let schemasCache: Record<string, MultiStepFormSchema> = {};

/**
 * Connector information derived from API templates.
 */
export interface ConnectorInfo {
  name: string;
  displayName: string;
  category: ConnectorCategory;
}

/**
 * Build the schema registry from ListTemplates API response.
 * For source drivers: uses the {driver}-duckdb template's json_schema.
 * For OLAP engines: uses the OLAP connector template's json_schema.
 *
 * The json_schema from the API is identical to the former TypeScript schemas;
 * we inject `title` from the template's display_name so existing consumers
 * that read schema.title continue to work.
 */
function buildSchemaRegistry(
  templates: V1Template[],
): Record<string, MultiStepFormSchema> {
  const schemas: Record<string, MultiStepFormSchema> = {};

  for (const t of templates) {
    if (!t.jsonSchema || !t.driver) continue;

    const key = t.driver;

    // Sources: pick the DuckDB-model template (has full connector+source form schema)
    if (SOURCES_SET.has(key) && t.olap === "duckdb") {
      schemas[key] = {
        ...t.jsonSchema,
        title: t.displayName,
      } as unknown as MultiStepFormSchema;
      continue;
    }

    // OLAP engines: pick the OLAP connector template (no olap set)
    if (OLAP_SET.has(key) && (!t.olap || t.olap === "")) {
      schemas[key] = {
        ...t.jsonSchema,
        title: t.displayName,
      } as unknown as MultiStepFormSchema;
    }
  }

  return schemas;
}

/**
 * Create a TanStack Query that fetches templates and provides the schema registry.
 * Call this once in AddDataModal (the modal entry point).
 *
 * Returns a reactive `connectors` store (for the connector grid) and the
 * underlying query (for loading/error state).
 * As a side effect, populates the module-level schemasCache so that
 * getConnectorSchema() works synchronously in child components.
 */
export function createConnectorSchemas() {
  const query = createRuntimeServiceListTemplates();

  const connectors: Readable<ConnectorInfo[]> = derived(query, ($q) => {
    if (!$q.data?.templates) return [];

    const schemas = buildSchemaRegistry($q.data.templates);

    // Populate module-level cache for sync access by child components
    schemasCache = schemas;

    return [...SOURCES, ...OLAP_ENGINES]
      .filter((name) => schemas[name]?.["x-category"])
      .map((name) => ({
        name,
        displayName: schemas[name].title ?? name,
        category: schemas[name]["x-category"] as ConnectorCategory,
      }));
  });

  return { query, connectors };
}

/**
 * Directly populate the schema cache.
 * Used in tests and for non-component contexts where the TanStack Query
 * is not available.
 */
export function populateSchemaCache(
  schemas: Record<string, MultiStepFormSchema>,
) {
  schemasCache = schemas;
}

/**
 * Get the schema for a connector by name.
 * Reads from the cache populated by createConnectorSchemas().
 */
export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema = schemasCache[connectorName];
  return schema?.properties ? schema : null;
}

/**
 * Get the backend driver name for a given schema name.
 * With API-driven schemas, the key is already the driver name.
 * Falls back to x-driver if specified (for future use).
 */
export function getBackendConnectorName(schemaName: string): string {
  const schema = getConnectorSchema(schemaName);
  return schema?.["x-driver"] ?? schemaName;
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
 * SQL stores and warehouses can browse tables and write custom queries.
 */
export function hasExplorerStep(schema: MultiStepFormSchema | null): boolean {
  const category = schema?.["x-category"];
  return category === "sqlStore" || category === "warehouse";
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
 * Get the form width CSS class for a connector's add data modal.
 * Some connectors with templates or more content use a wider form.
 */
export function getFormWidth(schema: MultiStepFormSchema | null): string {
  return schema?.["x-form-width"] === "wide" ? "max-w-5xl" : "max-w-4xl";
}
