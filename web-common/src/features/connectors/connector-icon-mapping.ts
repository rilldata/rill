import AmazonS3Icon from "../../components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "../../components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "../../components/icons/connectors/ApachePinotIcon.svelte";
import ClaudeIcon from "../../components/icons/connectors/ClaudeIcon.svelte";
import ClickHouseIcon from "../../components/icons/connectors/ClickHouseIcon.svelte";
import ClickHouseCloudIcon from "../../components/icons/connectors/ClickHouseCloudIcon.svelte";
import ApacheIcebergIcon from "../../components/icons/connectors/ApacheIcebergIcon.svelte";
import DeltaLakeIcon from "../../components/icons/connectors/DeltaLakeIcon.svelte";
import DuckDbIcon from "../../components/icons/connectors/DuckDBIcon.svelte";
import GeminiIcon from "../../components/icons/connectors/GeminiIcon.svelte";
import GoogleBigQueryIcon from "../../components/icons/connectors/GoogleBigQueryIcon.svelte";
import AthenaIcon from "../../components/icons/connectors/AthenaIcon.svelte";
import OpenAIIcon from "../../components/icons/connectors/OpenAIIcon.svelte";
import PostgresIcon from "../../components/icons/connectors/PostgresIcon.svelte";
import MySqlIcon from "../../components/icons/connectors/MySqlIcon.svelte";
import MotherDuckIcon from "../../components/icons/connectors/MotherDuckIcon.svelte";
import RedshiftIcon from "../../components/icons/connectors/RedshiftIcon.svelte";
import SnowflakeIcon from "../../components/icons/connectors/SnowflakeIcon.svelte";
import SalesforceIcon from "../../components/icons/connectors/SalesforceIcon.svelte";
import StarRocksIcon from "../../components/icons/connectors/StarRocksIcon.svelte";
import MicrosoftAzureBlobStorageIcon from "@rilldata/web-common/components/icons/connectors/MicrosoftAzureBlobStorageIcon.svelte";
import SupabaseIcon from "../../components/icons/connectors/SupabaseIcon.svelte";
import { File } from "lucide-svelte";
import GoogleCloudStorageIcon from "@rilldata/web-common/components/icons/connectors/GoogleCloudStorageIcon.svelte";
import HTTPSIcon from "@rilldata/web-common/components/icons/connectors/HTTPSIcon.svelte";

export const connectorIconMapping = {
  athena: AthenaIcon,
  azure: MicrosoftAzureBlobStorageIcon,
  bigquery: GoogleBigQueryIcon,
  claude: ClaudeIcon,
  clickhouse: ClickHouseIcon,
  delta: DeltaLakeIcon,
  clickhousecloud: ClickHouseCloudIcon,
  gemini: GeminiIcon,
  motherduck: MotherDuckIcon,
  druid: ApacheDruidIcon,
  duckdb: DuckDbIcon,
  gcs: GoogleCloudStorageIcon,
  iceberg: ApacheIcebergIcon,
  mysql: MySqlIcon,
  openai: OpenAIIcon,
  pinot: ApachePinotIcon,
  postgres: PostgresIcon,
  redshift: RedshiftIcon,
  s3: AmazonS3Icon,
  salesforce: SalesforceIcon,
  snowflake: SnowflakeIcon,
  sqlite: RedshiftIcon,
  starrocks: StarRocksIcon,
  supabase: SupabaseIcon,
  local_file: File,
  https: HTTPSIcon,
};

export const connectorClassMapping = {
  local_file: "text-slate-300",
  https: "text-slate-300",
};

export const connectorLabelMapping = {
  duckdb: "DuckDB",
  clickhouse: "ClickHouse",
  motherduck: "MotherDuck",
  s3: "S3",
  gcs: "GCS",
  snowflake: "Snowflake",
  druid: "Druid",
  starrocks: "StarRocks",
};
