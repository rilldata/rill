import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createRuntimeServiceAnalyzeConnectors,
  type V1AnalyzedConnector,
} from "@rilldata/web-common/runtime-client";
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

function groupBySchema(connectors: V1AnalyzedConnector[]) {
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

export function inferSchemaForConnector(connector: V1AnalyzedConnector) {
  // TODO: some schema will share driver name, differentiate them.
  return connector.driver?.name ?? "";
}
