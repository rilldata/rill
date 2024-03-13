import { useChart } from "@rilldata/web-common/features/charts/selectors";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  createRuntimeServiceGenerateChartSpec,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { Document } from "yaml";

export function createChartGenerator(instanceId: string, chart: string) {
  const generateVegaConfig = createRuntimeServiceGenerateChartSpec();
  const chartQuery = useChart(instanceId, chart);

  return async (prompt: string) => {
    const resp = await get(generateVegaConfig).mutateAsync({
      instanceId,
      data: {
        prompt,
        chart,
      },
    });
    const chartSpec = get(chartQuery).data?.chart?.spec;
    await runtimeServicePutFile(
      instanceId,
      getFileAPIPathFromNameAndType(chart, EntityType.Chart),
      {
        blob: getChartYaml(resp.vegaLiteSpec ?? "", {
          sql:
            chartSpec?.resolver === "SQL"
              ? chartSpec?.resolverProperties?.sql
              : undefined,
          metricsSql:
            chartSpec?.resolver === "MetricsSQL"
              ? chartSpec?.resolverProperties?.sql
              : undefined,
          api: chartSpec?.resolverProperties?.api,
        }),
      },
    );
  };
}

export function createFullChartGenerator(instanceId: string) {
  const generateVegaConfig = createRuntimeServiceGenerateChartSpec();

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
        blob: getChartYaml(resp.vegaLiteSpec ?? "", { sql: resp.sql ?? "" }),
      },
    );
  };
}

export function getChartYaml(
  vegaLite: string,
  {
    sql,
    metricsSql,
    api,
  }: {
    sql?: string;
    metricsSql?: string;
    api?: string;
  },
) {
  const doc = new Document();
  doc.set("kind", "chart");

  // TODO: more fields
  if (sql) {
    doc.set("data", { sql });
  } else if (metricsSql) {
    doc.set("data", { metrics_sql: metricsSql });
  } else if (api) {
    doc.set("data", { api });
  }

  doc.set(
    "vega_lite",
    JSON.stringify(JSON.parse(vegaLite), null, 2).replace(/^/gm, "  "),
  );

  return doc.toString();
}
