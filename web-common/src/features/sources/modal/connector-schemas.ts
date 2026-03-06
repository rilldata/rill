import type { ComponentType, SvelteComponent } from "svelte";
import {
  createRuntimeServiceGetInstance,
  createRuntimeServiceListTemplates,
  type V1Connector,
  type V1Template,
} from "../../../runtime-client";
import type { RuntimeClient } from "../../../runtime-client/v2";
import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "../../templates/schemas/types";
import type { ConnectorStep } from "./connectorStepStore";
import { SOURCES, OLAP_ENGINES } from "./constants";
import { derived, type Readable } from "svelte/store";
import { clearOlapCache } from "./generate-template";

export type ConnectorIcon = ComponentType<SvelteComponent>;

/**
 * Auto-discovered icon components from the connectors icon directory.
 * Keyed by filename (without .svelte extension), matching the x-icon / x-small-icon
 * strings in template JSON definitions.
 * To add a new icon: just drop a .svelte file in the connectors icon directory
 * and reference its name in the template JSON.
 */
const iconModules = import.meta.glob<{ default: ConnectorIcon }>(
  "../../../components/icons/connectors/*.svelte",
  { eager: true },
);

const ICON_COMPONENTS: Record<string, ConnectorIcon> = {};
for (const [path, mod] of Object.entries(iconModules)) {
  const name = path.match(/\/([^/]+)\.svelte$/)?.[1];
  if (name) ICON_COMPONENTS[name] = mod.default;
}

const SOURCES_SET = new Set(SOURCES);
const OLAP_SET = new Set(OLAP_ENGINES);

// Module-level caches populated when the TanStack Query resolves.
// Safe because AddDataModal (the entry point) subscribes to the query
// and renders the connector grid (step 1) first; by the time step 2
// needs getConnectorSchema(), the cache is populated.
let schemasCache: Record<string, MultiStepFormSchema> = {};
let iconsCache: Record<string, { icon?: string; smallIcon?: string }> = {};

/**
 * Connector information derived from API templates.
 */
export interface ConnectorInfo {
  name: string;
  displayName: string;
  category: ConnectorCategory;
  docsUrl?: string;
}

/**
 * Resolve the OLAP template suffix from the instance's OLAP connector.
 * Looks up the connector by name in both the instance's connectors and
 * project connectors lists to get the actual driver type. Project-defined
 * connectors (e.g. from connectors/*.yaml) live in projectConnectors.
 * Only ClickHouse has its own model templates; everything else uses DuckDB.
 */
export function normalizeOlapForTemplate(
  olapConnectorName: string,
  connectors?: V1Connector[],
  projectConnectors?: V1Connector[],
): string {
  const allConnectors = [...(connectors ?? []), ...(projectConnectors ?? [])];
  const connector = allConnectors.find((c) => c.name === olapConnectorName);
  if (connector?.type === "clickhouse") return "clickhouse";
  return "duckdb";
}

/**
 * Build the schema registry from ListTemplates API response.
 * For source drivers: uses the {driver}-{olap} template's json_schema.
 * For OLAP engines: uses the OLAP connector template's json_schema.
 *
 * The json_schema from the API is identical to the former TypeScript schemas;
 * we inject `title` from the template's display_name so existing consumers
 * that read schema.title continue to work.
 */
interface RegistryEntry {
  schema: MultiStepFormSchema;
  docsUrl?: string;
  icon?: string;
  smallIcon?: string;
}

function buildSchemaRegistry(
  templates: V1Template[],
  olap: string,
): Record<string, RegistryEntry> {
  const entries: Record<string, RegistryEntry> = {};

  for (const t of templates) {
    if (!t.jsonSchema || !t.driver) continue;

    const key = t.driver;

    // Sources: pick the template matching the instance's OLAP
    if (SOURCES_SET.has(key) && t.olap === olap) {
      entries[key] = {
        schema: {
          ...t.jsonSchema,
          title: t.displayName,
        } as unknown as MultiStepFormSchema,
        docsUrl: t.docsUrl,
        icon: t.icon,
        smallIcon: t.smallIcon,
      };
      continue;
    }

    // OLAP engines: pick the OLAP connector template (no olap set)
    if (OLAP_SET.has(key) && (!t.olap || t.olap === "")) {
      entries[key] = {
        schema: {
          ...t.jsonSchema,
          title: t.displayName,
        } as unknown as MultiStepFormSchema,
        docsUrl: t.docsUrl,
        icon: t.icon,
        smallIcon: t.smallIcon,
      };
    }
  }

  return entries;
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
export function createConnectorSchemas(client: RuntimeClient) {
  const templatesQuery = createRuntimeServiceListTemplates(client, {});
  const instanceQuery = createRuntimeServiceGetInstance(client, {
    sensitive: true,
  });

  const connectors: Readable<ConnectorInfo[]> = derived(
    [templatesQuery, instanceQuery],
    ([$tq, $iq]) => {
      if (!$tq.data?.templates) return [];

      const olap = normalizeOlapForTemplate(
        $iq.data?.instance?.olapConnector ?? "duckdb",
        $iq.data?.instance?.connectors,
        $iq.data?.instance?.projectConnectors,
      );
      const entries = buildSchemaRegistry($tq.data.templates, olap);

      // Clear the OLAP cache used by generateTemplate so the next preview
      // picks up the current OLAP engine (e.g., after switching DuckDB ↔ ClickHouse).
      clearOlapCache();

      // Populate module-level caches for sync access by child components
      schemasCache = Object.fromEntries(
        Object.entries(entries).map(([k, v]) => [k, v.schema]),
      );
      iconsCache = Object.fromEntries(
        Object.entries(entries).map(([k, v]) => [
          k,
          { icon: v.icon, smallIcon: v.smallIcon },
        ]),
      );
      rebuildIconMaps();

      return [...SOURCES, ...OLAP_ENGINES]
        .filter((name) => entries[name]?.schema["x-category"])
        .map((name) => ({
          name,
          displayName: entries[name].schema.title ?? name,
          category: entries[name].schema["x-category"] as ConnectorCategory,
          docsUrl: entries[name].docsUrl,
        }));
    },
  );

  return { query: templatesQuery, connectors };
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
  rebuildIconMaps();
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
 * Get the schema name for a given backend driver name.
 * Reverse lookup: finds the schema name that maps to the given driver name.
 * For most connectors, driver name = schema name, but some have x-driver override.
 */
export function getSchemaNameFromDriver(driverName: string): string {
  // First, check if driver name matches a schema name directly
  if (driverName in schemasCache) {
    return driverName;
  }

  // If not, search for schema with matching x-driver
  for (const [schemaName, schema] of Object.entries(schemasCache)) {
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
 * Only true when the schema category allows it AND there are actual
 * explorer-step fields defined. This prevents connectors like SQLite
 * and Salesforce (which have no explorer fields) from triggering the
 * multi-step flow unnecessarily.
 */
export function hasExplorerStep(schema: MultiStepFormSchema | null): boolean {
  if (!schema?.properties) return false;
  const category = schema?.["x-category"];
  if (category !== "sqlStore" && category !== "warehouse") return false;
  return Object.values(schema.properties).some(
    (p) => p["x-step"] === "explorer",
  );
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

/**
 * Resolve an icon component by name from the auto-discovered icon modules.
 */
function resolveIcon(name: string | undefined): ConnectorIcon | undefined {
  if (!name) return undefined;
  return ICON_COMPONENTS[name];
}

/**
 * Get the full-size icon for a connector (used in add-data grid).
 */
export function getConnectorIcon(
  connectorName: string,
): ConnectorIcon | undefined {
  return resolveIcon(iconsCache[connectorName]?.icon);
}

/**
 * Get the small icon for a connector (used in nav, cards, dialog headers).
 * Falls back to the full-size icon when small_icon is not defined.
 */
export function getConnectorSmallIcon(
  connectorName: string,
): ConnectorIcon | undefined {
  const entry = iconsCache[connectorName];
  return resolveIcon(entry?.smallIcon) ?? resolveIcon(entry?.icon);
}

/**
 * Full-size icon components keyed by connector name.
 * Derived from schemas; populated when createConnectorSchemas() resolves.
 */
export let ICONS: Record<string, ConnectorIcon> = {};

/**
 * Small icon components keyed by connector name.
 * Derived from schemas; populated when createConnectorSchemas() resolves.
 * Falls back to x-icon when x-small-icon is not defined.
 * Includes clickhousecloud as a special case (distinct icon for managed ClickHouse).
 */
export let connectorIconMapping: Record<string, ConnectorIcon> = {};

function rebuildIconMaps() {
  const icons: Record<string, ConnectorIcon> = {};
  const smallIcons: Record<string, ConnectorIcon> = {};

  for (const [name, entry] of Object.entries(iconsCache)) {
    const icon = resolveIcon(entry.icon);
    if (icon) icons[name] = icon;

    const smallIcon = resolveIcon(entry.smallIcon) ?? icon;
    if (smallIcon) smallIcons[name] = smallIcon;
  }

  // ClickHouse Cloud uses a distinct icon determined by getConnectorIconKey()
  const chCloudIcon = ICON_COMPONENTS["ClickHouseCloudIcon"];
  if (chCloudIcon) smallIcons["clickhousecloud"] = chCloudIcon;

  ICONS = icons;
  connectorIconMapping = smallIcons;
}

export const connectorLabelMapping: Record<string, string> = {
  duckdb: "DuckDB",
  clickhouse: "ClickHouse",
};
