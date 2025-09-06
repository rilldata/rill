import AmazonAthena from "@rilldata/web-common/components/icons/connectors/AmazonAthena.svelte";
import AmazonRedshift from "@rilldata/web-common/components/icons/connectors/AmazonRedshift.svelte";
import MySQL from "@rilldata/web-common/components/icons/connectors/MySQL.svelte";
import AmazonS3 from "../../../components/icons/connectors/AmazonS3.svelte";
import ApacheDruid from "../../../components/icons/connectors/ApacheDruid.svelte";
import ApachePinot from "../../../components/icons/connectors/ApachePinot.svelte";
import ClickHouse from "../../../components/icons/connectors/ClickHouse.svelte";
import DuckDB from "../../../components/icons/connectors/DuckDB.svelte";
import GoogleBigQuery from "../../../components/icons/connectors/GoogleBigQuery.svelte";
import GoogleCloudStorage from "../../../components/icons/connectors/GoogleCloudStorage.svelte";
import Https from "../../../components/icons/connectors/HTTPS.svelte";
import LocalFile from "../../../components/icons/connectors/LocalFile.svelte";
import MicrosoftAzureBlobStorage from "../../../components/icons/connectors/MicrosoftAzureBlobStorage.svelte";
import MotherDuck from "../../../components/icons/connectors/MotherDuck.svelte";
import Postgres from "../../../components/icons/connectors/Postgres.svelte";
import Salesforce from "../../../components/icons/connectors/Salesforce.svelte";
import Snowflake from "../../../components/icons/connectors/Snowflake.svelte";
import SQLite from "../../../components/icons/connectors/SQLite.svelte";

export const ICONS = {
  gcs: GoogleCloudStorage,
  s3: AmazonS3,
  azure: MicrosoftAzureBlobStorage,
  bigquery: GoogleBigQuery,
  athena: AmazonAthena,
  redshift: AmazonRedshift,
  duckdb: DuckDB,
  motherduck: MotherDuck,
  postgres: Postgres,
  mysql: MySQL,
  sqlite: SQLite,
  snowflake: Snowflake,
  salesforce: Salesforce,
  local_file: LocalFile,
  https: Https,
  clickhouse: ClickHouse,
  druid: ApacheDruid,
  pinot: ApachePinot,
};
