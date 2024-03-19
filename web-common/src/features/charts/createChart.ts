import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { getFileAPIPathFromNameAndType } from "../entity-management/entity-mappers";

export async function createChart(instanceId: string, newChartName: string) {
  await runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(newChartName, EntityType.Chart),
    {
      blob: `kind: chart
data:
  metrics_sql: |
    SELECT advertiser_name, AGGREGATE(measure_2) as measure_2
    FROM Bids_Sample_Dash
    GROUP BY advertiser_name
    LIMIT 20

vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": {"name": "table"},
    "mark": "bar",
    "width": "container",
    "encoding": {
      "x": {"field": "advertiser_name", "type": "nominal"},
      "y": {"field": "measure_2", "type": "quantitative"}
    }
  }`,
      createOnly: true,
    },
  );
}
