import { goto } from "$app/navigation";
import {
  getChartYaml,
  parseChartYaml,
} from "@rilldata/web-common/features/charts/chartYaml";
import {
  chartPromptsStore,
  ChartPromptStatus,
} from "@rilldata/web-common/features/charts/prompt/chartPrompt";
import { useChart } from "@rilldata/web-common/features/charts/selectors";
import {
  getFileAPIPathFromNameAndType,
  removeLeadingSlash,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  createRuntimeServiceGenerateRenderer,
  createRuntimeServiceGenerateResolver,
  createRuntimeServiceGetFile,
  runtimeServicePutFile,
  V1ComponentSpec,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function createChartGenerator(
  instanceId: string,
  chart: string,
  filePath: string,
) {
  const generateVegaConfig = createRuntimeServiceGenerateRenderer();
  const chartQuery = useChart(instanceId, chart);
  const chartContent = createRuntimeServiceGetFile(instanceId, filePath);
  // TODO: update for new API

  return async (prompt: string) => {
    try {
      const [resolver, resolverProperties] = tryParseChart(
        get(chartQuery).data?.component?.spec,
        get(chartContent).data?.blob,
      );
      chartPromptsStore.startPrompt(chart, chart, prompt);
      const resp = await get(generateVegaConfig).mutateAsync({
        instanceId,
        data: {
          prompt,
          resolver,
          resolverProperties,
        },
      });
      chartPromptsStore.updatePromptStatus(chart, ChartPromptStatus.Idle);
      await runtimeServicePutFile(instanceId, removeLeadingSlash(filePath), {
        blob: getChartYaml(
          resp.rendererProperties?.spec,
          resolver,
          resolverProperties,
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
  const generateVegaConfig = createRuntimeServiceGenerateRenderer();

  return async (
    prompt: string,
    {
      table,
      connector,
      metricsView,
    }: { table?: string; connector?: string; metricsView?: string },
    newChartName: string,
  ) => {
    const filePath = getFileAPIPathFromNameAndType(
      newChartName,
      EntityType.Chart,
    );
    try {
      // add an empty chart
      await runtimeServicePutFile(instanceId, filePath, {
        blob: `kind: chart`,
      });
      chartPromptsStore.startPrompt(
        (table || metricsView) ?? "",
        newChartName,
        prompt,
      );
      await goto(`/files//${filePath}`);
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
      await runtimeServicePutFile(instanceId, filePath, {
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
      await runtimeServicePutFile(instanceId, filePath, {
        blob: getChartYaml(
          resp.rendererProperties?.spec,
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

function tryParseChart(
  chartSpec: V1ComponentSpec | undefined,
  chartContent: string | undefined,
): [resolver: string, resolverProperties: Record<string, string>] {
  if (!chartSpec?.resolver && chartContent) {
    try {
      chartSpec = parseChartYaml(chartContent);
    } catch (err) {
      throw new Error(
        "Failed to parse yaml. Please fix it before trying to generate chart spec.",
      );
    }
  }
  if (chartSpec?.resolver) {
    return [chartSpec.resolver, chartSpec.resolverProperties ?? {}];
  }
  throw new Error("Chart is invalid");
}
