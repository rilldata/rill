import type { MultiStepFormSchema } from "../../templates/schemas/types";
import { athenaSchema } from "../../templates/schemas/athena";
import { azureSchema } from "../../templates/schemas/azure";
import { bigquerySchema } from "../../templates/schemas/bigquery";
import { clickhouseSchema } from "../../templates/schemas/clickhouse";
import { gcsSchema } from "../../templates/schemas/gcs";
import { mysqlSchema } from "../../templates/schemas/mysql";
import { postgresSchema } from "../../templates/schemas/postgres";
import { redshiftSchema } from "../../templates/schemas/redshift";
import { salesforceSchema } from "../../templates/schemas/salesforce";
import { snowflakeSchema } from "../../templates/schemas/snowflake";
import { sqliteSchema } from "../../templates/schemas/sqlite";
import { httpsSchema } from "../../templates/schemas/https";
import { s3Schema } from "../../templates/schemas/s3";

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
  https: httpsSchema,
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
};

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
  if (!schema?.properties) return null;
  return schema;
}
