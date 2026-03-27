import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "@rilldata/web-common/features/connectors/selectors.ts";
import { derived } from "svelte/store";
import { connectors } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import type { AddDataConfig } from "@rilldata/web-common/features/add-data/steps/types.ts";

export function getSupportedConnectorInfos(
  runtimeClient: RuntimeClient,
  config: AddDataConfig,
) {
  const isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(runtimeClient);

  return derived(
    isModelingSupportedForDefaultOlapDriver,
    (isModellingSupportedResp) => {
      return connectors
        .filter(
          (c) =>
            (config.importOnly ? true : c.name !== "duckdb") &&
            c.category !== "ai" &&
            (isModellingSupportedResp.data || c.category === "olap"),
        )
        .sort((a, b) => {
          if (a.name === "https" || a.name === "local_file") return 1;
          if (b.name === "https" || b.name === "local_file") return -1;
          return a.displayName.localeCompare(b.displayName);
        });
    },
  );
}

const TopConnectors = ["clickhouse", "gcs", "s3", "snowflake"];
const TopConnectorsWithoutModeling = [
  "clickhouse",
  "motherduck",
  "druid",
  "starrocks",
];
export function getSupportedTopConnectors(runtimeClient: RuntimeClient) {
  const isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(runtimeClient);

  return derived(
    isModelingSupportedForDefaultOlapDriver,
    (isModellingSupportedResp) => {
      return isModellingSupportedResp.data
        ? TopConnectors
        : TopConnectorsWithoutModeling;
    },
  );
}
