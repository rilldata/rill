import { SnowflakeIcon } from "lucide-svelte";
import AmazonS3Icon from "../../components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "../../components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "../../components/icons/connectors/ApachePinotIcon.svelte";
import ClickHouseIcon from "../../components/icons/connectors/ClickHouseIcon.svelte";
import DuckDbIcon from "../../components/icons/connectors/DuckDBIcon.svelte";
import GoogleBigQueryIcon from "../../components/icons/connectors/GoogleBigQueryIcon.svelte";

export const connectorIconMapping = {
  // TODO: athena
  // TODO: azure
  bigquery: GoogleBigQueryIcon,
  clickhouse: ClickHouseIcon,
  druid: ApacheDruidIcon,
  duckdb: DuckDbIcon,
  // TODO: gcs
  // TODO: mysql
  pinot: ApachePinotIcon,
  // TODO: postgres
  // TODO: redshift
  s3: AmazonS3Icon,
  snowflake: SnowflakeIcon,
};
