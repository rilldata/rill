import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client";
import type { AnalyzedConnector } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
import { structToJson } from "@rilldata/web-common/lib/proto-utils";
import { isDuckLakeConnector } from "@rilldata/web-common/features/connectors/connectors-utils";
import { connectorInfoMap } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";

type ConnectorForSchema = {
  displayName: string;
  connector: string;
  schema: string;
};

export function getConnectorsWithImportSupport(runtimeClient: RuntimeClient) {
  return createRuntimeServiceAnalyzeConnectors(
    runtimeClient,
    {},
    {
      query: {
        select: (data) => {
          const connectorsWithImportSupport = data?.connectors?.filter(
            (c) =>
              c.driver?.implementsObjectStore ||
              c.driver?.implementsSqlStore ||
              c.driver?.implementsWarehouse,
          );
          return groupBySchema(connectorsWithImportSupport ?? []);
        },
      },
    },
  );
}

export function getConnectorsWithMetricsViewSupport(
  runtimeClient: RuntimeClient,
) {
  return createRuntimeServiceAnalyzeConnectors(
    runtimeClient,
    {},
    {
      query: {
        select: (data) => {
          const connectorsWithMetricsViewSupport = data?.connectors?.filter(
            (c) => c.driver?.implementsOlap,
          );
          return groupBySchema(connectorsWithMetricsViewSupport ?? []);
        },
      },
    },
  );
}

function groupBySchema(connectors: AnalyzedConnector[]) {
  const seenDriver = new Set<string>();
  const connectorsForSchema: ConnectorForSchema[] = [];
  connectors.forEach((connector) => {
    const driverName = inferSchemaForConnector(connector);
    if (seenDriver.has(driverName)) return;

    const info = connectorInfoMap.get(driverName);
    connectorsForSchema.push({
      displayName: info?.displayName ?? driverName,
      connector: connector.name!,
      schema: driverName,
    });
    seenDriver.add(driverName);
  });
  return connectorsForSchema;
}

export function inferSchemaForConnector(connector: AnalyzedConnector) {
  if (!connector.driver?.name) return "";
  // TODO: some schema will share driver name, differentiate them.
  const driverName = connector.driver.name;
  switch (driverName) {
    case "duckdb":
      if (String(structToJson(connector.config)?.path ?? "").startsWith("md:"))
        return "motherduck";
      if (isDuckLakeConnector(connector)) return "ducklake";
      break;
  }

  return driverName;
}
