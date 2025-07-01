export const CLICKHOUSE_DEFAULTS = {
  "rill-managed": {},
  "self-managed": {
    host: {
      value: "",
      placeholder:
        "your-instance.clickhouse.cloud or your.clickhouse.server.com",
      hint: "Your ClickHouse hostname (e.g., your-instance.clickhouse.cloud or your-server.com)",
    },
    port: {
      value: "9000",
      placeholder: "9000",
      hint: "Default port is 9000 for native protocol. Also commonly used: 8443 for ClickHouse Cloud (HTTPS), 8123 for HTTP",
    },
    username: {
      value: "",
      placeholder: "default",
      hint: "Username for authentication",
    },
    password: {
      value: "",
      placeholder: "Database password",
      hint: "Password to your database",
    },
    cluster: {
      value: "",
      placeholder: "Cluster name",
      hint: "Cluster name (required for some self-hosted ClickHouse setups)",
    },
    ssl: { value: true, hint: "Enable SSL for secure connections" },
  },
};

export type ClickHouseConnectorType = keyof typeof CLICKHOUSE_DEFAULTS;

// FIXME: rill-managed is only available locally
// https://docs.rilldata.com/reference/olap-engines/clickhouse#configuring-rill-cloud
export const CONNECTOR_TYPE_OPTIONS: {
  value: ClickHouseConnectorType;
  label: string;
}[] = [
  { value: "self-managed", label: "Self-managed ClickHouse" },
  { value: "rill-managed", label: "Rill-managed ClickHouse" },
];
