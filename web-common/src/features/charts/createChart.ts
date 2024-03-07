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
    metrics_view: Bids_Sample_Dash
    measures:
      - name: measure_2
    dimensions:
      - name: advertiser_name
    timeRange:
      start: '2023-03-07T00:30:00.000Z'
      end: '2023-03-08T00:30:00.000Z'
    sort:
    - desc: true
      name: measure_2
    limit: '20'
    offset: '0'

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
