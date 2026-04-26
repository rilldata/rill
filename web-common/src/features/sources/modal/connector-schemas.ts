import type { V1ConnectorDriver } from "../../../runtime-client";
import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "../../templates/schemas/types";
import type { ConnectorStep } from "./connectorStepStore";
import { athenaSchema } from "../../templates/schemas/athena";
import { azureSchema } from "../../templates/schemas/azure";
import { bigquerySchema } from "../../templates/schemas/bigquery";
import { claudeSchema } from "../../templates/schemas/claude";
import { clickhouseSchema } from "../../templates/schemas/clickhouse";
import { gcsSchema } from "../../templates/schemas/gcs";
import { geminiSchema } from "../../templates/schemas/gemini";
import { mysqlSchema } from "../../templates/schemas/mysql";
import { openaiSchema } from "../../templates/schemas/openai";
import { postgresSchema } from "../../templates/schemas/postgres";
import { redshiftSchema } from "../../templates/schemas/redshift";
import { salesforceSchema } from "../../templates/schemas/salesforce";
import { snowflakeSchema } from "../../templates/schemas/snowflake";
import { sqliteSchema } from "../../templates/schemas/sqlite";
import { localFileSchema } from "../../templates/schemas/local_file";
import { duckdbSchema } from "../../templates/schemas/duckdb";
import { ducklakeSchema } from "../../templates/schemas/ducklake";
import { deltaSchema } from "../../templates/schemas/delta";
import { httpsSchema } from "../../templates/schemas/https";
import { icebergSchema } from "../../templates/schemas/iceberg";
import { motherduckSchema } from "../../templates/schemas/motherduck";
import { druidSchema } from "../../templates/schemas/druid";
import { pinotSchema } from "../../templates/schemas/pinot";
import { s3Schema } from "../../templates/schemas/s3";
import { starrocksSchema } from "../../templates/schemas/starrocks";
import { supabaseSchema } from "../../templates/schemas/supabase";
import { SOURCES, OLAP_ENGINES, AI_CONNECTORS } from "./constants";
import { connectorKeywordMapping } from "@rilldata/web-common/features/connectors/connector-metadata.ts";

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  athena: athenaSchema,
  bigquery: bigquerySchema,
  clickhouse: clickhouseSchema,
  mysql: mysqlSchema,
  postgres: postgresSchema,
  redshift: redshiftSchema,
  salesforce: salesforceSchema,
  snowflake: snowflakeSchema,
  sqlite: sqliteSchema,
  motherduck: motherduckSchema,
  duckdb: duckdbSchema,
  ducklake: ducklakeSchema,
  druid: druidSchema,
  pinot: pinotSchema,
  starrocks: starrocksSchema,
  supabase: supabaseSchema,
  local_file: localFileSchema,
  https: httpsSchema,
  s3: s3Schema,
  gcs: gcsSchema,
  iceberg: icebergSchema,
  azure: azureSchema,
  delta: deltaSchema,
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
 * All connectors enumerated from JSON schemas, sorted by display order.
 */
export const connectors: ConnectorInfo[] = [
  ...SOURCES,
  ...OLAP_ENGINES,
  ...AI_CONNECTORS,
]
  .filter((name) => multiStepFormSchemas[name]?.["x-category"])
  .map((name) => {
    const schema = multiStepFormSchemas[name];
    return {
      name,
      displayName: schema.title ?? name,
      category: schema["x-category"] as ConnectorCategory,
      keywords: connectorKeywordMapping[name] ?? [],
    };
  });
/**
 * Map of connector names to ConnectorInfo objects.
 * We need connector info by name in a lot of places, so we have a map to optimize lookups.
 */
export const connectorInfoMap = new Map<string, ConnectorInfo>(
  connectors.map((connector) => [connector.name, connector]),
);

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema = multiStepFormSchemas[connectorName];
  return schema?.properties ? schema : null;
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
