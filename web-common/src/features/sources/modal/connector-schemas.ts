import type { MultiStepFormSchema } from "../../templates/schemas/types";
import { athenaSchema } from "../../templates/schemas/athena";
import { azureSchema } from "../../templates/schemas/azure";
import { bigquerySchema } from "../../templates/schemas/bigquery";
import { druidSchema } from "../../templates/schemas/druid";
import { duckdbSchema } from "../../templates/schemas/duckdb";
import { gcsSchema } from "../../templates/schemas/gcs";
import { httpsSchema } from "../../templates/schemas/https";
import { motherduckSchema } from "../../templates/schemas/motherduck";
import { mysqlSchema } from "../../templates/schemas/mysql";
import { pinotSchema } from "../../templates/schemas/pinot";
import { postgresSchema } from "../../templates/schemas/postgres";
import { redshiftSchema } from "../../templates/schemas/redshift";
import { s3Schema } from "../../templates/schemas/s3";
import { snowflakeSchema } from "../../templates/schemas/snowflake";

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
  https: httpsSchema,
  postgres: postgresSchema,
  mysql: mysqlSchema,
  snowflake: snowflakeSchema,
  bigquery: bigquerySchema,
  redshift: redshiftSchema,
  athena: athenaSchema,
  duckdb: duckdbSchema,
  motherduck: motherduckSchema,
  druid: druidSchema,
  pinot: pinotSchema,
};

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
  if (!schema?.properties) return null;
  return schema;
}

export function isStepMatch(
  schema: MultiStepFormSchema | null,
  key: string,
  step?: "connector" | "source" | string,
): boolean {
  if (!schema?.properties) return false;
  const prop = schema.properties[key];
  if (!prop) return false;
  if (!step) return true;
  const propStep = prop["x-step"];
  if (!propStep) return true;
  return propStep === step;
}
