import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createRuntimeServiceGetInstance,
  createRuntimeServiceListTemplates,
  getRuntimeServiceListTemplatesQueryOptions,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver } from "@rilldata/web-common/features/connectors/selectors.ts";
import { connectorKeywordMapping } from "@rilldata/web-common/features/connectors/connector-metadata.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  registerTemplateSchema,
  type ConnectorInfo,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import { setOlapCache } from "@rilldata/web-common/features/sources/modal/generate-template.ts";
import type { AddDataConfig } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
import type {
  ConnectorCategory,
  MultiStepFormSchema,
} from "@rilldata/web-common/features/templates/schemas/types";
import type { Template as V1Template } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";

/**
 * Register schemas from `ListTemplates` responses so getConnectorSchema and
 * connectorInfoMap resolve drivers that only exist as templates (e.g. kafka,
 * hudi when ClickHouse is the OLAP).
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
  fallbackCategory: ConnectorCategory,
): ConnectorInfo {
  const driver = t.driver ?? t.name ?? "";
  const schema = t.jsonSchema as Record<string, unknown> | undefined;
  const category =
    (schema?.["x-category"] as ConnectorCategory | undefined) ??
    fallbackCategory;
  return {
    name: driver,
    displayName: t.displayName ?? t.name ?? driver,
    category,
    keywords: connectorKeywordMapping[driver] ?? [],
  };
}

export function getSupportedConnectorInfos(
  runtimeClient: RuntimeClient,
  config: AddDataConfig,
) {
  const instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });

  // Source templates re-fetch whenever the instance OLAP changes.
  const sourceQueryOptions = derived([instanceQuery], ([$instance]) => {
    const olap = $instance.data?.instance?.olapConnector || "";
    if (olap) setOlapCache(runtimeClient.instanceId, olap);
    return getRuntimeServiceListTemplatesQueryOptions(
      runtimeClient,
      { tags: ["source", olap] },
      { query: { enabled: !!olap } },
    );
  });
  const sourceTemplates = createQuery(sourceQueryOptions);
  const olapTemplates = createRuntimeServiceListTemplates(runtimeClient, {
    tags: ["olap"],
  });

  return derived([sourceTemplates, olapTemplates], ([$sources, $olap]) => {
    const sourceList = ($sources.data?.templates ?? []) as V1Template[];
    const olapList = ($olap.data?.templates ?? []) as V1Template[];
    registerTemplatesIfNeeded([...sourceList, ...olapList]);

    const sources = sourceList.map((t) =>
      templateToConnectorInfo(t, "sourceOnly" as ConnectorCategory),
    );
    const olaps = olapList.map((t) =>
      templateToConnectorInfo(t, "olap" as ConnectorCategory),
    );

    // Deduplicate by driver name; source entries take priority over OLAP
    // entries when a connector appears in both lists (e.g. ClickHouse).
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
