export type ClickHouseConnectorType = "self-managed" | "rill-managed";

export const CONNECTOR_TYPE_OPTIONS: {
  value: ClickHouseConnectorType;
  label: string;
}[] = [
  { value: "self-managed", label: "Self-managed ClickHouse" },
  { value: "rill-managed", label: "Rill-managed ClickHouse" },
];
