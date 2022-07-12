import { useInlineTestServer } from "../utils/useInlineTestServer";
import request from "supertest";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  TimeSeriesResponse,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";
import { useBasicMetricsDefinition } from "../utils/metrics-definition-helpers";
import {
  assertTimeSeries,
  assertTimeSeriesMeasureRange,
} from "../utils/time-series-helpers";
import { MetricsExploreTestData } from "../data/MetricsExplore.data";

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
      .get(`/api/metrics/${metricsDef.id}/time-range`)
      .set("Accept", "application/json");
    const timeRange = resp.body.data as TimeSeriesTimeRange;
    expect(timeRange.interval).toBe("1 day");
    expect(timeRange.start).not.toBeUndefined();
    expect(timeRange.end).not.toBeUndefined();
  });

  for (const MetricsExploreTest of MetricsExploreTestData) {
    it(`Should return time series for ${MetricsExploreTest.title}`, async () => {
      // select measures based on index passed or default to all measures
      const requestMeasures = MetricsExploreTest.measures
        ? MetricsExploreTest.measures.map((index) => measures[index])
        : measures;
      const resp = await request(inlineServer.app)
        .post(`/api/metrics/${metricsDef.id}/time-series`)
        .send({
          measures: requestMeasures,
          filters: MetricsExploreTest.filters ?? {},
          ...(MetricsExploreTest.timeRange
            ? { timeRange: MetricsExploreTest.timeRange }
            : {}),
        })
        .set("Accept", "application/json");

      const timeSeries = resp.body as TimeSeriesResponse;

      assertTimeSeries(
        timeSeries,
        MetricsExploreTest.previewRollupInterval,
        requestMeasures.map((measure) => measure.sqlName)
      );
      if (MetricsExploreTest.measureRanges) {
        assertTimeSeriesMeasureRange(
          timeSeries,
          MetricsExploreTest.measureRanges
        );
      }
    });
  }
});
