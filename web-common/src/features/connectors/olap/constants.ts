export const CLICKHOUSE_DEFAULTS = {
  "rill-managed": {},
  "self-managed": {
    host: {
      value: "",
      placeholder:
        "your-instance.clickhouse.cloud or your.clickhouse.server.com",
    },
    port: { value: "9000", placeholder: "9000" },
    username: { value: "", placeholder: "default" },
    password: { value: "", placeholder: "Your ClickHouse database password" },
    ssl: { value: true },
  },
};

export type ClickHouseDeploymentType = keyof typeof CLICKHOUSE_DEFAULTS;

export const DEPLOYMENT_TYPE_OPTIONS: {
  value: ClickHouseDeploymentType;
  label: string;
}[] = [
  { value: "self-managed", label: "Self-managed ClickHouse" },
  { value: "rill-managed", label: "Rill-managed ClickHouse" },
];
