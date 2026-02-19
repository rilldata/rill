import type { V1ConnectorDriver } from "../../../runtime-client";
import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "../../templates/schemas/types";
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
import { httpsSchema } from "../../templates/schemas/https";
import { motherduckSchema } from "../../templates/schemas/motherduck";
import { druidSchema } from "../../templates/schemas/druid";
import { pinotSchema } from "../../templates/schemas/pinot";
import { s3Schema } from "../../templates/schemas/s3";
import { starrocksSchema } from "../../templates/schemas/starrocks";
import { SOURCES, OLAP_ENGINES, AI_CONNECTORS } from "./constants";

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
  druid: druidSchema,
  pinot: pinotSchema,
  starrocks: starrocksSchema,
  local_file: localFileSchema,
  https: httpsSchema,
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
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
    };
  });

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
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
 * Determine if a connector is an AI connector (Claude, OpenAI, Gemini).
 * AI connectors save directly without testing the connection.
 */
export function isAiConnector(schema: MultiStepFormSchema | null): boolean {
  return schema?.["x-category"] === "ai";
}

/**
 * Build a V1ConnectorDriver-compatible object from a schema name.
 * Shared by AddDataModal.toConnectorDriver and addSourceModal.openForConnector.
 */
export function toConnectorDriver(
  schemaName: string,
): V1ConnectorDriver | null {
  const schema = getConnectorSchema(schemaName);
  if (!schema) return null;
  const category = schema["x-category"];
  const backendName = getBackendConnectorName(schemaName);
  const docsCategory =
    category === "ai"
      ? "services"
      : category === "olap"
        ? "olap"
        : "data-source";
  return {
    name: backendName,
    displayName: schema.title ?? schemaName,
    docsUrl: `https://docs.rilldata.com/developers/build/connectors/${docsCategory}/${backendName}`,
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
 * Get the form width CSS class for a connector's add data modal.
 * Some connectors with templates or more content use a wider form.
 */
export function getFormWidth(schema: MultiStepFormSchema | null): string {
  return schema?.["x-form-width"] === "wide" ? "max-w-5xl" : "max-w-4xl";
}
