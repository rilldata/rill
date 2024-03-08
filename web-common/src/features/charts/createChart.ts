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
  name: MetricsViewAggregation
  args:
    metrics_view: foo

vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": {
      "values": [
        {"category":"A", "group": "x", "value":0.1},
        {"category":"A", "group": "y", "value":0.6},
        {"category":"A", "group": "z", "value":0.9},
        {"category":"B", "group": "x", "value":0.7},
        {"category":"B", "group": "y", "value":0.2},
        {"category":"B", "group": "z", "value":1.1},
        {"category":"C", "group": "x", "value":0.6},
        {"category":"C", "group": "y", "value":0.1},
        {"category":"C", "group": "z", "value":0.2}
      ]
    },
    "mark": "bar",
    "encoding": {
      "x": {"field": "category"},
      "y": {"field": "value", "type": "quantitative"},
      "xOffset": {"field": "group"},
      "color": {"field": "group"}
    }
  }`,
      createOnly: true,
    },
  );
}
