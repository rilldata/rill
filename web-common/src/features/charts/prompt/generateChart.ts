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

  return async (table: string, prompt: string) => {
    const resp = await get(generateVegaConfig).mutateAsync({
      instanceId,
      data: {
        prompt,
        table,
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

function getChartYaml(sql: string, vegaLite: string) {
  return `kind: chart
data:
  metrics_sql: |
${sql.replace(/^/gm, "    ")}

vega_lite: |
${vegaLite.replace(/^/gm, "  ")}
`;
}
