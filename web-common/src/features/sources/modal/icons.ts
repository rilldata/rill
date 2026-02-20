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
import ClickHouseCloud from "../../../components/icons/connectors/ClickHouseCloud.svelte";
import Staging from "../../../components/icons/connectors/Staging.svelte";
import StarRocks from "../../../components/icons/connectors/StarRocks.svelte";

// Icon-only versions (square, with size prop) for use in compact UI like rich selects
import AmazonAthenaIcon from "@rilldata/web-common/components/icons/connectors/AthenaIcon.svelte";
import AmazonRedshiftIcon from "@rilldata/web-common/components/icons/connectors/RedshiftIcon.svelte";
import MySQLIcon from "@rilldata/web-common/components/icons/connectors/MySqlIcon.svelte";
import AmazonS3Icon from "@rilldata/web-common/components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "@rilldata/web-common/components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "@rilldata/web-common/components/icons/connectors/ApachePinotIcon.svelte";
import ClickHouseIcon from "@rilldata/web-common/components/icons/connectors/ClickHouseIcon.svelte";
import DuckDBIcon from "@rilldata/web-common/components/icons/connectors/DuckDBIcon.svelte";
import GoogleBigQueryIcon from "@rilldata/web-common/components/icons/connectors/GoogleBigQueryIcon.svelte";
import MotherDuckIcon from "@rilldata/web-common/components/icons/connectors/MotherDuckIcon.svelte";
import PostgresIcon from "@rilldata/web-common/components/icons/connectors/PostgresIcon.svelte";
import SalesforceIcon from "@rilldata/web-common/components/icons/connectors/SalesforceIcon.svelte";
import SnowflakeIcon from "@rilldata/web-common/components/icons/connectors/SnowflakeIcon.svelte";
import ClickHouseCloudIcon from "@rilldata/web-common/components/icons/connectors/ClickHouseCloudIcon.svelte";
import StarRocksIcon from "@rilldata/web-common/components/icons/connectors/StarRocksIcon.svelte";

/** Full connector logos (with wordmarks) for the Add Data modal */
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
  clickhousecloud: ClickHouseCloud,
  staging: Staging,
  druid: ApacheDruid,
  pinot: ApachePinot,
  starrocks: StarRocks,
};

/** Square icon-only versions keyed by driver name, for compact rich selects */
export const DRIVER_ICONS = {
  gcs: GoogleCloudStorage,
  s3: AmazonS3Icon,
  azure: MicrosoftAzureBlobStorage,
  bigquery: GoogleBigQueryIcon,
  athena: AmazonAthenaIcon,
  redshift: AmazonRedshiftIcon,
  duckdb: DuckDBIcon,
  motherduck: MotherDuckIcon,
  postgres: PostgresIcon,
  mysql: MySQLIcon,
  sqlite: SQLite,
  snowflake: SnowflakeIcon,
  salesforce: SalesforceIcon,
  local_file: LocalFile,
  https: Https,
  clickhouse: ClickHouseIcon,
  clickhousecloud: ClickHouseCloudIcon,
  staging: Staging,
  druid: ApacheDruidIcon,
  pinot: ApachePinotIcon,
  starrocks: StarRocksIcon,
};
