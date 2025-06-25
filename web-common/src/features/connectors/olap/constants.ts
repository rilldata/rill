export const CLICKHOUSE_DEFAULTS = {
  "rill-managed": {
    host: { value: "", placeholder: "Managed by Rill" },
    port: { value: "9000", placeholder: "9000" },
    username: { value: "", placeholder: "default" },
    password: { value: "", placeholder: "Your ClickHouse password" },
    ssl: { value: true },
  },
  "self-managed": {
    host: { value: "", placeholder: "your-clickhouse-server.com" },
    port: { value: "9000", placeholder: "9000" },
    username: { value: "", placeholder: "default" },
    password: { value: "", placeholder: "Your ClickHouse password" },
    ssl: { value: true },
  },
};

export type ClickHouseDeploymentType = keyof typeof CLICKHOUSE_DEFAULTS;

export const DEPLOYMENT_TYPE_OPTIONS: {
  value: ClickHouseDeploymentType;
  label: string;
}[] = [
  { value: "rill-managed", label: "Rill Managed" },
  { value: "self-managed", label: "Self Managed" },
];
