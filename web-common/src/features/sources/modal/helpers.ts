import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { ClickHouseConnectorType } from "./constants";

export function hasOnlyDsn(
  connector: V1ConnectorDriver | undefined,
  isConnectorForm: boolean,
): boolean {
  if (!isConnectorForm) return false;
  const props = connector?.configProperties ?? [];
  const hasDsn = props.some((p) => p.key === "dsn");
  const hasOthers = props.some((p) => p.key !== "dsn");
  return hasDsn && !hasOthers;
}

export function applyClickHouseCloudRequirements(
  connectorName: string | undefined,
  connectorType: ClickHouseConnectorType,
  values: Record<string, unknown>,
): Record<string, unknown> {
  if (connectorName === "clickhouse" && connectorType === "clickhouse-cloud") {
    return { ...values, ssl: true, port: "8443" } as Record<string, unknown>;
  }
  return values;
}


