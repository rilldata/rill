import AmazonS3Icon from "../../components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "../../components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "../../components/icons/connectors/ApachePinotIcon.svelte";
import AzureBlobStorageIcon from "../../components/icons/connectors/AzureBlobStorageIcon.svelte";
import ClaudeIcon from "../../components/icons/connectors/ClaudeIcon.svelte";
import ClickHouseIcon from "../../components/icons/connectors/ClickHouseIcon.svelte";
import ClickHouseCloudIcon from "../../components/icons/connectors/ClickHouseCloudIcon.svelte";
import ApacheIcebergIcon from "../../components/icons/connectors/ApacheIcebergIcon.svelte";
import DeltaLakeIcon from "../../components/icons/connectors/DeltaLakeIcon.svelte";
import DuckDbIcon from "../../components/icons/connectors/DuckDBIcon.svelte";
import DuckLakeIcon from "../../components/icons/connectors/DuckLakeIcon.svelte";
import GeminiIcon from "../../components/icons/connectors/GeminiIcon.svelte";
import GoogleBigQueryIcon from "../../components/icons/connectors/GoogleBigQueryIcon.svelte";
import GoogleCloudStorageIcon from "../../components/icons/connectors/GoogleCloudStorageIcon.svelte";
import HttpsIcon from "../../components/icons/connectors/HttpsIcon.svelte";
import AthenaIcon from "../../components/icons/connectors/AthenaIcon.svelte";
import LocalFileIcon from "../../components/icons/connectors/LocalFileIcon.svelte";
import PostgresIcon from "../../components/icons/connectors/PostgresIcon.svelte";
import MySqlIcon from "../../components/icons/connectors/MySqlIcon.svelte";
import MotherDuckIcon from "../../components/icons/connectors/MotherDuckIcon.svelte";
import OpenAIIcon from "../../components/icons/connectors/OpenAIIcon.svelte";
import RedshiftIcon from "../../components/icons/connectors/RedshiftIcon.svelte";
import SnowflakeIcon from "../../components/icons/connectors/SnowflakeIcon.svelte";
import SalesforceIcon from "../../components/icons/connectors/SalesforceIcon.svelte";
import StarRocksIcon from "../../components/icons/connectors/StarRocksIcon.svelte";
import SQLiteIcon from "../../components/icons/connectors/SQLiteIcon.svelte";
import SupabaseIcon from "../../components/icons/connectors/SupabaseIcon.svelte";

export const connectorIconMapping = {
  athena: AthenaIcon,
  azure: AzureBlobStorageIcon,
  bigquery: GoogleBigQueryIcon,
  claude: ClaudeIcon,
  clickhouse: ClickHouseIcon,
  delta: DeltaLakeIcon,
  clickhousecloud: ClickHouseCloudIcon,
  druid: ApacheDruidIcon,
  duckdb: DuckDbIcon,
  ducklake: DuckLakeIcon,
  gcs: GoogleCloudStorageIcon,
  https: HttpsIcon,
  iceberg: ApacheIcebergIcon,
  local_file: LocalFileIcon,
  gemini: GeminiIcon,
  motherduck: MotherDuckIcon,
  mysql: MySqlIcon,
  openai: OpenAIIcon,
  pinot: ApachePinotIcon,
  postgres: PostgresIcon,
  redshift: RedshiftIcon,
  s3: AmazonS3Icon,
  salesforce: SalesforceIcon,
  snowflake: SnowflakeIcon,
  sqlite: SQLiteIcon,
  starrocks: StarRocksIcon,
  supabase: SupabaseIcon,
};

export const connectorClassMapping = {
  local_file: "text-slate-300",
  https: "text-slate-300",
};

export const connectorLabelMapping = {
  duckdb: "DuckDB",
  ducklake: "DuckLake",
  clickhouse: "ClickHouse",
  motherduck: "MotherDuck",
  s3: "S3",
  gcs: "GCS",
  snowflake: "Snowflake",
  druid: "Druid",
  starrocks: "StarRocks",
};

export const connectorKeywordMapping = {
  athena: ["aws", "amazon", "serverless", "presto"],
  azure: ["object storage", "cloud storage"],
  bigquery: ["google", "gcp", "data warehouse"],
  clickhouse: ["olap", "columnar", "real-time analytics"],
  deltalake: ["databricks", "lakehouse", "spark", "parquet"],
  druid: ["apache", "olap", "time series"],
  gcs: ["bucket", "object storage"],
  iceberg: ["apache", "table format", "lakehouse", "parquet"],
  motherduck: ["duckdb", "serverless"],
  ducklake: [
    "duckdb",
    "lakehouse",
    "catalog",
    "parquet",
    "iceberg alternative",
  ],
  mysql: ["relational", "rdbms", "database"],
  pinot: ["apache", "olap", "real-time analytics"],
  postgres: ["relational", "rdbms", "database"],
  redshift: ["aws", "amazon", "data warehouse"],
  s3: ["aws", "amazon", "bucket", "object storage"],
  snowflake: ["data warehouse", "cloud"],
  sqlite: ["embedded", "lightweight", "local database"],
  starrocks: ["olap", "columnar", "mpp"],
  supabase: ["postgres", "baas", "firebase alternative"],
  https: ["url", "api", "remote", "endpoint"],
  local_file: ["upload", "csv", "parquet", "json", "import"],
};
