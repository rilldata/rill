export const CLICKHOUSE_DEFAULTS = {
  cloud: {
    host: { value: "", placeholder: "your-instance.clickhouse.cloud" },
    port: { value: "9000", placeholder: "9000" },
    username: { value: "", placeholder: "default" },
    password: { value: "", placeholder: "Your ClickHouse Cloud password" },
    ssl: { value: true },
  },
  "self-hosted": {
    host: { value: "", placeholder: "your-clickhouse-server.com" },
    port: { value: "9000", placeholder: "9000" },
    username: { value: "", placeholder: "default" },
    password: { value: "", placeholder: "Your ClickHouse password" },
    ssl: { value: true },
  },
  local: {
    host: { value: "localhost", placeholder: "localhost" },
    port: { value: "9000", placeholder: "9000" },
    username: { value: "", placeholder: "default" },
    password: { value: "", placeholder: "Your ClickHouse password" },
    ssl: { value: false },
  },
};

export type ClickHouseDeploymentType = keyof typeof CLICKHOUSE_DEFAULTS;

export const DEPLOYMENT_TYPE_OPTIONS: {
  value: ClickHouseDeploymentType;
  label: string;
}[] = [
  { value: "cloud", label: "ClickHouse Cloud" },
  { value: "self-hosted", label: "Self-Hosted ClickHouse" },
  { value: "local", label: "Local Development" },
];
