import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type {
  TimeSeriesResponse,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";
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

  it("Should return estimated time", async () => {
    const resp = await request(inlineServer.app)
      .get(`/api/metrics/${metricsDef.id}/all-time-range`)
      .set("Accept", "application/json");
    const timeRange = resp.body.data as TimeSeriesTimeRange;
    expect(timeRange.interval).toBe("1 day");
    expect(timeRange.start).not.toBeUndefined();
    expect(timeRange.end).not.toBeUndefined();
  });

  for (const MetricsExplorerTest of MetricsExplorerTestData) {
    it(`Should return time series for ${MetricsExplorerTest.title}`, async () => {
      // select measures based on index passed or default to all measures
      const requestMeasures = MetricsExplorerTest.measures
        ? MetricsExplorerTest.measures.map((index) => measures[index])
        : measures;
      const resp = await request(inlineServer.app)
        .post(`/api/metrics/${metricsDef.id}/time-series`)
        .send({
          measures: requestMeasures,
          filters: MetricsExplorerTest.filters ?? {},
          ...(MetricsExplorerTest.timeRange
            ? { timeRange: MetricsExplorerTest.timeRange }
            : {}),
        })
        .set("Accept", "application/json");

      const timeSeries = resp.body as TimeSeriesResponse;

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
