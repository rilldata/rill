import { useChart } from "@rilldata/web-common/features/charts/selectors";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  createRuntimeServiceGenerateChartSpec,
  createRuntimeServiceGenerateResolver,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { Document } from "yaml";

export function createChartGenerator(instanceId: string, chart: string) {
  const generateVegaConfig = createRuntimeServiceGenerateChartSpec();
  const chartQuery = useChart(instanceId, chart);

  return async (prompt: string) => {
    const chartSpec = get(chartQuery).data?.chart?.spec;
    const resp = await get(generateVegaConfig).mutateAsync({
      instanceId,
      data: {
        prompt,
        resolver: chartSpec?.resolver,
        resolverProperties: chartSpec?.resolverProperties,
      },
    });
    await runtimeServicePutFile(
      instanceId,
      getFileAPIPathFromNameAndType(chart, EntityType.Chart),
      {
        blob: getChartYaml(
          resp.vegaLiteSpec,
          chartSpec?.resolver,
          chartSpec?.resolverProperties,
        ),
      },
    );
  };
}

export function createFullChartGenerator(instanceId: string) {
  const generateResolver = createRuntimeServiceGenerateResolver();
  const generateVegaConfig = createRuntimeServiceGenerateChartSpec();

  return async (
    prompt: string,
    {
      table,
      connector,
      metricsView,
    }: { table?: string; connector?: string; metricsView?: string },
    newChartName: string,
  ) => {
    const resolverResp = await get(generateResolver).mutateAsync({
      instanceId,
      data: {
        table,
        connector,
        metricsView,
        prompt,
      },
    });
    const resp = await get(generateVegaConfig).mutateAsync({
      instanceId,
      data: {
        prompt,
        resolver: resolverResp.resolver,
        resolverProperties: resolverResp.resolverProperties,
      },
    });
    await runtimeServicePutFile(
      instanceId,
      getFileAPIPathFromNameAndType(newChartName, EntityType.Chart),
      {
        blob: getChartYaml(
          resp.vegaLiteSpec,
          resolverResp.resolver,
          resolverResp.resolverProperties,
        ),
      },
    );
  };
}

export function getChartYaml(
  vegaLite: string | undefined,
  resolver: string | undefined,
  resolverProperties: Record<string, any> | undefined,
) {
  const doc = new Document();
  doc.set("kind", "chart");

  // TODO: more fields from resolverProperties
  if (resolver === "SQL") {
    doc.set("data", { sql: (resolverProperties?.sql as string) ?? "" });
  } else if (resolver === "MetricsSQL") {
    doc.set("data", { metrics_sql: (resolverProperties?.sql as string) ?? "" });
  } else if (resolver === "API") {
    doc.set("data", { api: (resolverProperties?.api as string) ?? "" });
  }

  doc.set(
    "vega_lite",
    JSON.stringify(JSON.parse(vegaLite ?? "{}"), null, 2).replace(/^/gm, "  "),
  );

  return doc.toString();
}
