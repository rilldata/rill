import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createRuntimeServiceGetInstance,
  createRuntimeServiceListTemplates,
  getRuntimeServiceListTemplatesQueryOptions,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "@rilldata/web-common/features/connectors/selectors";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  registerTemplateSchema,
  type ConnectorInfo,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import type { AddDataConfig } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
import type { V1Template } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types";

/**
 * Register template schemas so that downstream lookups
 * (getConnectorSchema, getConnectorDriverForSchema, connectorInfoMap)
 * work for connectors that only exist in templates (e.g. kafka, hudi for ClickHouse).
 * Also overrides static schemas when a template provides a different x-category
 * (e.g. GCS is "objectStore" statically but "sourceOnly" for ClickHouse OLAP).
 */
function registerTemplatesIfNeeded(templates: V1Template[]) {
  for (const t of templates) {
    const driver = t.driver ?? t.name ?? "";
    const templateName = t.name ?? "";
    if (!driver || !t.jsonSchema) continue;
    registerTemplateSchema(
      driver,
      templateName,
      t.jsonSchema as unknown as MultiStepFormSchema,
      t.displayName ?? t.name ?? driver,
    );
  }
}

function templateToConnectorInfo(
  t: V1Template,
  category: string,
): ConnectorInfo {
  return {
    name: t.driver ?? t.name ?? "",
    displayName: t.displayName ?? t.name ?? "",
    category: category as any,
  };
}

export function getSupportedConnectorInfos(
  runtimeClient: RuntimeClient,
  config: AddDataConfig,
) {
  const instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });

  // Build reactive query options that re-fetch when the OLAP driver changes
  const sourceQueryOptions = derived([instanceQuery], ([$instance]) => {
    const olap = $instance.data?.instance?.olapConnector || "";
    return getRuntimeServiceListTemplatesQueryOptions(
      runtimeClient,
      { tags: ["source", olap] },
      { query: { enabled: !!olap } },
    );
  });
  const sourceTemplates = createQuery(sourceQueryOptions);

  // OLAP templates (static, no dependency on instance)
  const olapTemplates = createRuntimeServiceListTemplates(runtimeClient, {
    tags: ["olap"],
  });

  return derived([sourceTemplates, olapTemplates], ([$sources, $olap]) => {
    const allTemplates = [
      ...($sources.data?.templates ?? []),
      ...($olap.data?.templates ?? []),
    ];

    // Register template schemas so form lookups work for all connectors
    registerTemplatesIfNeeded(allTemplates);

    const sources = ($sources.data?.templates ?? []).map((t) =>
      templateToConnectorInfo(t, "source"),
    );
    const olaps = ($olap.data?.templates ?? []).map((t) =>
      templateToConnectorInfo(t, "olap"),
    );

    // Deduplicate by name: source entries take priority over OLAP entries
    // (e.g. "clickhouse" can appear as both a source and OLAP template)
    const seen = new Set<string>();
    const merged: ConnectorInfo[] = [];
    for (const c of [...sources, ...olaps]) {
      if (!seen.has(c.name)) {
        seen.add(c.name);
        merged.push(c);
      }
    }

    return merged
      .filter(
        (c) =>
          (config.importOnly ? true : c.name !== "duckdb") &&
          c.category !== "ai",
      )
      .sort((a, b) => {
        if (a.name === "https" || a.name === "local_file") return 1;
        if (b.name === "https" || b.name === "local_file") return -1;
        return a.displayName.localeCompare(b.displayName);
      });
  });
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
