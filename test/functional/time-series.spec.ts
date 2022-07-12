import {
  useInlineTestServer,
  useTestModel,
  useTestTables,
} from "../utils/useInlineTestServer";
import request from "supertest";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  TimeSeriesResponse,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";
import { PreviewRollupInterval } from "$lib/duckdb-data-types";
import {
  getMetricsDefinition,
  setupMeasures,
  useMetricsDefinition,
} from "../utils/metrics-definition-helpers";
import {
  assertTimeSeries,
  assertTimeSeriesMeasureRange,
  getTimeRange,
  TimeSeriesMeasureRange,
} from "../utils/time-series-helpers";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { START_DATE } from "../data/generator/data-constants";

describe("TimeSeries", () => {
  const { inlineServer } = useInlineTestServer(8082);
  const AdEventsName = "AdEvents";

  let metricsDef: MetricsDefinitionEntity;
  let measures: Array<MeasureDefinitionEntity>;

  useTestTables(inlineServer);
  useTestModel(
    inlineServer,
    `select
    bid.*, imp.user_id, imp.city, imp.country
    from AdBids bid join AdImpressions imp on bid.id = imp.id`,
    AdEventsName
  );
  useMetricsDefinition(inlineServer, AdEventsName, AdEventsName, "timestamp");
  setupMeasures(inlineServer, AdEventsName, "impressions", [
    { id: "", expression: "avg(bid_price)", sqlName: "bid_price" },
  ]);
  getMetricsDefinition(
    inlineServer,
    AdEventsName,
    (selMetricsDef, selMeasures) => {
      metricsDef = selMetricsDef;
      measures = selMeasures;
    }
  );

  it("Should return estimated time", async () => {
    const resp = await request(inlineServer.app)
      .get(`/api/metrics/${metricsDef.id}/time-range`)
      .set("Accept", "application/json");
    const timeRange = resp.body.data as TimeSeriesTimeRange;
    expect(timeRange.interval).toBe("1 day");
    expect(timeRange.start).not.toBeUndefined();
    expect(timeRange.end).not.toBeUndefined();
  });

  const TimeSeriesTestData: Array<{
    title: string;
    measures?: Array<number>;
    filters?: ActiveValues;
    previewRollupInterval: PreviewRollupInterval;
    timeRange?: TimeSeriesTimeRange;
    measureRanges?: Array<TimeSeriesMeasureRange>;
  }> = [
    {
      title: "Should return a basic time series",
      previewRollupInterval: PreviewRollupInterval.day,
    },
    {
      title: "Should return a time series by month",
      timeRange: getTimeRange(PreviewRollupInterval.month),
      previewRollupInterval: PreviewRollupInterval.month,
    },
    {
      title: "Should return a time series with filters",
      timeRange: getTimeRange(PreviewRollupInterval.month),
      previewRollupInterval: PreviewRollupInterval.month,
      filters: {
        domain: [["sports.yahoo.com", true]],
      },
      measureRanges: [
        { impressions: [3500, 4500], bid_price: [3, 4] },
        { impressions: [750, 1250], bid_price: [1, 2] },
        { impressions: [750, 1250], bid_price: [1, 2] },
      ],
    },
    {
      title:
        "Should return a time series with filters and time range and single measure",
      timeRange: getTimeRange(PreviewRollupInterval.month, "2022-02-01"),
      previewRollupInterval: PreviewRollupInterval.month,
      filters: {
        publisher: [["Yahoo", false]],
      },
      measures: [1],
      measureRanges: [{ bid_price: [2.5, 3] }, { bid_price: [3, 3.5] }],
    },
    {
      title: "Should return a time series with missing start time",
      timeRange: getTimeRange(
        PreviewRollupInterval.month,
        "2021-11-01",
        "2022-02-28"
      ),
      previewRollupInterval: PreviewRollupInterval.month,
      measures: [1],
      measureRanges: [
        { bid_price: [0, 0] },
        { bid_price: [0, 0] },
        { bid_price: [2.75, 3.25] },
        { bid_price: [2.75, 3.25] },
      ],
    },
  ];

  for (const TimeSeriesTest of TimeSeriesTestData) {
    it(TimeSeriesTest.title, async () => {
      // select measures based on index passed or default to all measures
      const requestMeasures = TimeSeriesTest.measures
        ? TimeSeriesTest.measures.map((index) => measures[index])
        : measures;
      const resp = await request(inlineServer.app)
        .post(`/api/metrics/${metricsDef.id}/time-series`)
        .send({
          measures: requestMeasures,
          filters: TimeSeriesTest.filters ?? {},
          ...(TimeSeriesTest.timeRange
            ? { timeRange: TimeSeriesTest.timeRange }
            : {}),
        })
        .set("Accept", "application/json");

      const timeSeries = resp.body as TimeSeriesResponse;

      assertTimeSeries(
        timeSeries,
        TimeSeriesTest.previewRollupInterval,
        requestMeasures.map((measure) => measure.sqlName)
      );
      if (TimeSeriesTest.measureRanges) {
        assertTimeSeriesMeasureRange(timeSeries, TimeSeriesTest.measureRanges);
      }
    });
  }
});
