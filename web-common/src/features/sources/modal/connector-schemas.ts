import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "../../templates/schemas/types";
import { athenaSchema } from "../../templates/schemas/athena";
import { azureSchema } from "../../templates/schemas/azure";
import { bigquerySchema } from "../../templates/schemas/bigquery";
import { clickhouseSchema } from "../../templates/schemas/clickhouse";
import { clickhousecloudSchema } from "../../templates/schemas/clickhousecloud";
import { gcsSchema } from "../../templates/schemas/gcs";
import { mysqlSchema } from "../../templates/schemas/mysql";
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
import { SOURCES, OLAP_ENGINES } from "./constants";

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  athena: athenaSchema,
  bigquery: bigquerySchema,
  clickhouse: clickhouseSchema,
  clickhousecloud: clickhousecloudSchema,
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
  local_file: localFileSchema,
  https: httpsSchema,
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
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
export const connectors: ConnectorInfo[] = [...SOURCES, ...OLAP_ENGINES]
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
 * Object store connectors (S3, GCS, Azure) require separate auth and source steps.
 */
export function isMultiStepConnector(
  schema: MultiStepFormSchema | null,
): boolean {
  return schema?.["x-category"] === "objectStore";
}

/**
 * Determine if a connector supports explorer mode (SQL query interface).
 * SQL stores, warehouses, and OLAP engines can browse tables and write custom queries.
 */
export function hasExplorerStep(schema: MultiStepFormSchema | null): boolean {
  const category = schema?.["x-category"];
  return category === "sqlStore" || category === "warehouse" || category === "olap";
}

// Form height CSS classes for modals
export const FORM_HEIGHT_TALL = "max-h-[38.5rem] min-h-[38.5rem]";
export const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";

/**
 * Get the form height CSS class for a connector's add data modal.
 * Some connectors with more fields use a taller form.
 */
export function getFormHeight(schema: MultiStepFormSchema | null): string {
  return schema?.["x-form-height"] === "tall"
    ? FORM_HEIGHT_TALL
    : FORM_HEIGHT_DEFAULT;
}
