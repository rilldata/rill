import AmazonS3Icon from "../../components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "../../components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "../../components/icons/connectors/ApachePinotIcon.svelte";
import ClickHouseIcon from "../../components/icons/connectors/ClickHouseIcon.svelte";
import DuckDbIcon from "../../components/icons/connectors/DuckDBIcon.svelte";
import GoogleBigQueryIcon from "../../components/icons/connectors/GoogleBigQueryIcon.svelte";
import AthenaIcon from "../../components/icons/connectors/AthenaIcon.svelte";
import PostgresIcon from "../../components/icons/connectors/PostgresIcon.svelte";
import MySqlIcon from "../../components/icons/connectors/MySqlIcon.svelte";
import MotherDuckIcon from "../../components/icons/connectors/MotherDuckIcon.svelte";
import RedshiftIcon from "../../components/icons/connectors/RedshiftIcon.svelte";
import SnowflakeIcon from "../../components/icons/connectors/SnowflakeIcon.svelte";

export const connectorIconMapping = {
  athena: AthenaIcon,
  bigquery: GoogleBigQueryIcon,
  clickhouse: ClickHouseIcon,
  motherduck: MotherDuckIcon,
  druid: ApacheDruidIcon,
  duckdb: DuckDbIcon,
  mysql: MySqlIcon,
  pinot: ApachePinotIcon,
  postgres: PostgresIcon,
  redshift: RedshiftIcon,
  s3: AmazonS3Icon,
  snowflake: SnowflakeIcon,
};
