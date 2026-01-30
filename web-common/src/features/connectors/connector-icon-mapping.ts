import AmazonS3Icon from "../../components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "../../components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "../../components/icons/connectors/ApachePinotIcon.svelte";
import ClickHouseIcon from "../../components/icons/connectors/ClickHouseIcon.svelte";
import DuckDbIcon from "../../components/icons/connectors/DuckDBIcon.svelte";
import GoogleBigQueryIcon from "../../components/icons/connectors/GoogleBigQueryIcon.svelte";
import GoogleCloudStorageIcon from "../../components/icons/connectors/GoogleCloudStorage.svelte";
import AzureBlobStorageIcon from "../../components/icons/connectors/MicrosoftAzureBlobStorage.svelte";
import AthenaIcon from "../../components/icons/connectors/AthenaIcon.svelte";
import PostgresIcon from "../../components/icons/connectors/PostgresIcon.svelte";
import MySqlIcon from "../../components/icons/connectors/MySqlIcon.svelte";
import MotherDuckIcon from "../../components/icons/connectors/MotherDuckIcon.svelte";
import RedshiftIcon from "../../components/icons/connectors/RedshiftIcon.svelte";
import SnowflakeIcon from "../../components/icons/connectors/SnowflakeIcon.svelte";
import ClickHouseCloudIcon from "../../components/icons/connectors/ClickHouseCloudIcon.svelte";
import StarRocksIcon from "../../components/icons/connectors/StarRocksIcon.svelte";

export const connectorIconMapping = {
  athena: AthenaIcon,
  azure: AzureBlobStorageIcon,
  bigquery: GoogleBigQueryIcon,
  clickhouse: ClickHouseIcon,
  clickhousecloud: ClickHouseCloudIcon,
  gcs: GoogleCloudStorageIcon,
  motherduck: MotherDuckIcon,
  druid: ApacheDruidIcon,
  duckdb: DuckDbIcon,
  mysql: MySqlIcon,
  pinot: ApachePinotIcon,
  postgres: PostgresIcon,
  redshift: RedshiftIcon,
  s3: AmazonS3Icon,
  snowflake: SnowflakeIcon,
  starrocks: StarRocksIcon,
};
