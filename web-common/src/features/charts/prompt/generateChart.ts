import { goto } from "$app/navigation";
import { useChart } from "@rilldata/web-common/features/charts/selectors";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  createRuntimeServiceGenerateChartFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function createChartGenerator(instanceId: string, chart: string) {
  const generateVegaConfig = createRuntimeServiceGenerateChartFile();
  const chartQuery = useChart(instanceId, chart);

  return async (prompt: string) => {
    const resp = await get(generateVegaConfig).mutateAsync({
      instanceId,
      data: {
        prompt,
        chart,
      },
    });
    await runtimeServicePutFile(
      instanceId,
      getFileAPIPathFromNameAndType(chart, EntityType.Chart),
      {
        blob: getChartYaml(
          get(chartQuery).data?.chart?.spec?.resolverProperties?.sql ?? "",
          resp.vegaLiteSpec ?? "",
        ),
      },
    );
  };
}

export function createFullChartGenerator(instanceId: string) {
  const generateVegaConfig = createRuntimeServiceGenerateChartFile();

  return async (table: string, prompt: string, newChartName: string) => {
    const resp = await get(generateVegaConfig).mutateAsync({
      instanceId,
      data: {
        table,
        prompt,
      },
    });
    await runtimeServicePutFile(
      instanceId,
      getFileAPIPathFromNameAndType(newChartName, EntityType.Chart),
      {
        blob: getChartYaml(resp.sql ?? "", resp.vegaLiteSpec ?? "", "sql"),
      },
    );
  };
}

function getChartYaml(
  sql: string,
  vegaLite: string,
  sqlField: "metrics_sql" | "sql" = "metrics_sql",
) {
  return `kind: chart
data:
  ${sqlField}: |
${sql.replace(/^/gm, "    ")}

vega_lite: |
${vegaLite.replace(/^/gm, "  ")}
`;
}
