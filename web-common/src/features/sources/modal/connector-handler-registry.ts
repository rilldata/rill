import { connectorHandlerRegistry } from "./connector-handlers";
import {
  StandardConnectorHandler,
  ClickHouseConnectorHandler,
} from "./connector-handler-implementations";

// Register all connector handlers
const connectors = [
  "s3",
  "gcs",
  "https",
  "duckdb",
  "motherduck",
  "sqlite",
  "bigquery",
  "azure",
  "postgres",
  "mysql",
  "redshift",
  "snowflake",
  "salesforce",
  "athena",
  "druid",
  "pinot",
];

// Register standard handlers for most connectors
connectors.forEach((connectorName) => {
  if (connectorName !== "clickhouse") {
    connectorHandlerRegistry.register(
      new StandardConnectorHandler(connectorName),
    );
  }
});

// Register ClickHouse handler
connectorHandlerRegistry.register(new ClickHouseConnectorHandler());
