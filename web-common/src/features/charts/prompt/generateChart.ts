import { goto } from "$app/navigation";
import { getChartYaml } from "@rilldata/web-common/features/charts/chartYaml";
import {
  chartPromptsStore,
  ChartPromptStatus,
} from "@rilldata/web-common/features/charts/prompt/chartPrompt";
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

export function createChartGenerator(instanceId: string, chart: string) {
  const generateVegaConfig = createRuntimeServiceGenerateChartSpec();
  const chartQuery = useChart(instanceId, chart);
  const chartPath = getFilePathFromNameAndType(chart, EntityType.Chart);

  return async (prompt: string) => {
    try {
      const chartSpec = get(chartQuery).data?.chart?.spec;
      chartPromptsStore.startPrompt(chart, chart, prompt);
      const resp = await get(generateVegaConfig).mutateAsync({
        instanceId,
        data: {
          prompt,
          resolver: chartSpec?.resolver,
          resolverProperties: chartSpec?.resolverProperties,
        },
      });
      chartPromptsStore.updatePromptStatus(chart, ChartPromptStatus.Idle);
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: getChartYaml(
          resp.vegaLiteSpec,
          chartSpec?.resolver,
          chartSpec?.resolverProperties,
        ),
      });
    } catch (e) {
      chartPromptsStore.setPromptError(
        chart,
        e.message ?? e.response.data.message,
      );
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
      chartPromptsStore.startPrompt(
        (table || metricsView) ?? "",
        newChartName,
        prompt,
      );
      await goto(getRouteFromName(newChartName, EntityType.Chart));
      const resolverResp = await get(generateResolver).mutateAsync({
        instanceId,
        data: {
          table,
          connector,
          metricsView,
          prompt,
        },
      });

      // add a chart with just the resolver
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: getChartYaml(
          "{}",
          resolverResp.resolver,
          resolverResp.resolverProperties,
        ),
      });
      chartPromptsStore.updatePromptStatus(
        newChartName,
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

      chartPromptsStore.updatePromptStatus(
        newChartName,
        ChartPromptStatus.Idle,
      );
      await runtimeServicePutFile(instanceId, chartPath, {
        blob: getChartYaml(
          resp.vegaLiteSpec,
          resolverResp.resolver,
          resolverResp.resolverProperties,
        ),
      });
    } catch (e) {
      chartPromptsStore.setPromptError(
        newChartName,
        e.message ?? e.response.data.message,
      );
    }
  };
}
