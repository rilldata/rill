import { goto } from "$app/navigation";
import {
  ChartPromptStatus,
  chartPromptStore,
} from "@rilldata/web-common/features/charts/prompt/chartPromptStatus";
import { useChart } from "@rilldata/web-common/features/charts/selectors";
import {
  getFilePathFromNameAndType,
  getRouteFromName,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
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
  const chartPath = getFilePathFromNameAndType(chart, EntityType.Chart);

  return async (prompt: string) => {
    try {
      const chartSpec = get(chartQuery).data?.chart?.spec;
      chartPromptStore.setStatus(
        chartPath,
        ChartPromptStatus.GeneratingChartSpec,
      );
      const resp = await get(generateVegaConfig).mutateAsync({
        instanceId,
        data: {
          prompt,
          resolver: chartSpec?.resolver,
          resolverProperties: chartSpec?.resolverProperties,
        },
      });
      chartPromptStore.deleteStatus(chartPath);
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: getChartYaml(
          resp.vegaLiteSpec,
          chartSpec?.resolver,
          chartSpec?.resolverProperties,
        ),
      });
    } catch (e) {
      chartPromptStore.deleteStatus(chartPath);
    }
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
    const chartPath = getFilePathFromNameAndType(
      newChartName,
      EntityType.Chart,
    );
    try {
      // add an empty chart
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: `kind: chart`,
      });
      chartPromptStore.setStatus(chartPath, ChartPromptStatus.GeneratingData);
      console.log(chartPath, "...1");
      await goto(getRouteFromName(newChartName, EntityType.Chart));
      console.log(chartPath, "...2");
      const resolverResp = await get(generateResolver).mutateAsync({
        instanceId,
        data: {
          table,
          connector,
          metricsView,
          prompt,
        },
      });
      console.log(chartPath, "...3");

      // add a chart with just the resolver
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: getChartYaml(
          "{}",
          resolverResp.resolver,
          resolverResp.resolverProperties,
        ),
      });
      console.log(chartPath, "...4");
      chartPromptStore.setStatus(
        chartPath,
        ChartPromptStatus.GeneratingChartSpec,
      );
      const resp = await get(generateVegaConfig).mutateAsync({
        instanceId,
        data: {
          prompt,
          resolver: resolverResp.resolver,
          resolverProperties: resolverResp.resolverProperties,
        },
      });
      console.log(chartPath, "...5");

      chartPromptStore.deleteStatus(chartPath);
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: getChartYaml(
          resp.vegaLiteSpec,
          resolverResp.resolver,
          resolverResp.resolverProperties,
        ),
      });
    } catch (e) {
      chartPromptStore.deleteStatus(chartPath);
    }
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
