import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type {
  MetricsViewTimeSeriesRequest,
  MetricsViewTimeSeriesResponse,
} from "$common/rill-developer-service/MetricsViewActions";
import request from "supertest";
import { MetricsExplorerTestData } from "../data/MetricsExplorer.data";
import { useBasicMetricsDefinition } from "../utils/metrics-definition-helpers";
import {
  assertTimeSeries,
  assertTimeSeriesMeasureRange,
} from "../utils/time-series-helpers";
import { useInlineTestServer } from "../utils/useInlineTestServer";

describe("TimeSeries", () => {
  const { inlineServer } = useInlineTestServer(8082);

  let metricsDef: MetricsDefinitionEntity;
  let measures: Array<MeasureDefinitionEntity>;
  useBasicMetricsDefinition(inlineServer, (selMetricsDef, selMeasures) => {
    metricsDef = selMetricsDef;
    measures = selMeasures;
  });

  for (const MetricsExplorerTest of MetricsExplorerTestData) {
    it(`Should return time series for ${MetricsExplorerTest.title}`, async () => {
      const requestMeasures = MetricsExplorerTest.measures
        ? MetricsExplorerTest.measures.map((index) => measures[index])
        : measures;
      const timeSeriesRequest: MetricsViewTimeSeriesRequest = {
        // select measures based on index passed or default to all measures
        measures: requestMeasures.map((measure) => measure.id),
        filter: MetricsExplorerTest.filters,
        time: {
          start: MetricsExplorerTest.timeRange?.start,
          end: MetricsExplorerTest.timeRange?.end,
          granularity: MetricsExplorerTest.timeRange?.interval,
        },
      };
      const resp = await request(inlineServer.app)
        .post(`/api/v1/metrics-views/${metricsDef.id}/timeseries`)
        .send(timeSeriesRequest)
        .set("Accept", "application/json");

      const timeSeries = resp.body as MetricsViewTimeSeriesResponse;

      assertTimeSeries(
        timeSeries,
        MetricsExplorerTest.previewRollupInterval,
        requestMeasures.map((measure) => measure.sqlName)
      );
      if (MetricsExplorerTest.measureRanges) {
        assertTimeSeriesMeasureRange(
          timeSeries,
          MetricsExplorerTest.measureRanges
        );
      }
    });
  }
});
