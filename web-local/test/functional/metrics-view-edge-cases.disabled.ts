import { describe, expect, it } from "@jest/globals";
import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { getFallbackMeasureName } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type {
  MetricsViewRequestFilter,
  MetricsViewTotalsRequest,
  MetricsViewTotalsResponse,
} from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import type { MetricsViewTopListRequest } from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import type {
  MetricsViewTimeSeriesRequest,
  MetricsViewTimeSeriesResponse,
} from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import type { LeaderboardValues } from "@rilldata/web-local/lib/application-state-stores/explorer-stores";
import { PreviewRollupInterval } from "@rilldata/web-local/lib/duckdb-data-types";
import axios from "axios";
import request from "supertest";
import {
  getMetricsDefinition,
  useMetricsDefinition,
} from "../utils/metrics-definition-helpers";
import { normaliseLeaderboardOrder } from "../utils/normaliseLeaderboardOrder";
import {
  assertBigNumber,
  assertTimeSeries,
  assertTimeSeriesMeasureRange,
} from "../utils/time-series-helpers";
import {
  useInlineTestServer,
  useTestModel,
  useTestTables,
} from "../utils/useInlineTestServer";

describe("Metrics View Edge Cases", () => {
  const { config, inlineServer } = useInlineTestServer(8084);
  useTestTables(inlineServer);

  describe("Metrics view from model with quotes", () => {
    const TableName = "QuotedColumns";
    let metricsDef: MetricsDefinitionEntity;
    let measureNames: Array<string>;
    let dimension: DimensionDefinitionEntity;
    let filter: MetricsViewRequestFilter;

    useTestModel(
      inlineServer,
      `select make_date(2022, month, "day"), replace(domain, '.com', '') from (
        select
          date_part('day', timestamp) as day,
          date_part('month', timestamp) as month, domain
        from AdBids group by day, month, domain
      )`,
      TableName
    );
    useMetricsDefinition(
      inlineServer,
      TableName,
      TableName,
      `make_date(2022, "month", "day")`
    );
    getMetricsDefinition(
      inlineServer,
      TableName,
      (selMetricsDef, selMeasures, selDimensions) => {
        metricsDef = selMetricsDef;
        measureNames = selMeasures.map((measure, idx) =>
          getFallbackMeasureName(idx, measure.sqlName)
        );
        dimension = selDimensions[0];
        filter = {
          include: [
            {
              name: dimension.dimensionColumn,
              in: [],
              like: ["%google%", "%yahoo%"],
            },
          ],
          exclude: [],
        };
      }
    );

    it("Time series", async () => {
      const timeSeriesRequest: MetricsViewTimeSeriesRequest = {
        // select measures based on index passed or default to all measures
        measures: measureNames,
        filter,
        time: {
          start: undefined,
          end: undefined,
          granularity: PreviewRollupInterval.month,
        },
      };
      const resp = await request(inlineServer.app)
        .post(`/api/v1/metrics-views/${metricsDef.id}/timeseries`)
        .send(timeSeriesRequest)
        .set("Accept", "application/json");

      const timeSeries = resp.body as MetricsViewTimeSeriesResponse;
      assertTimeSeries(timeSeries, PreviewRollupInterval.month, measureNames);
      assertTimeSeriesMeasureRange(timeSeries, [
        { measure_0: [100, 150] },
        { measure_0: [100, 150] },
        { measure_0: [100, 150] },
      ]);
    });

    it("Top list", async () => {
      const request: MetricsViewTopListRequest = {
        measures: measureNames,
        filter,
        time: {
          start: undefined,
          end: undefined,
        },
        limit: 15,
        offset: 0,
        sort: [{ name: measureNames[0], direction: "desc" }],
      };
      const resp = await axios.post(
        `${config.server.serverUrl}/api/v1/metrics-views/${metricsDef.id}/toplist/${dimension.id}`,
        request,
        { responseType: "json" }
      );

      expect(
        normaliseLeaderboardOrder(
          [
            {
              values: resp.data.data,
              dimensionName: dimension.dimensionColumn,
            } as LeaderboardValues,
          ],
          measureNames[0]
        )
      ).toStrictEqual([
        [
          `replace("domain", '.com', '')`,
          ["google", "news.google", "news.yahoo", "sports.yahoo"],
        ],
      ]);
    });

    it("Totals", async () => {
      const request: MetricsViewTotalsRequest = {
        // select measures based on index passed or default to all measures
        measures: measureNames,
        filter,
        time: {
          start: undefined,
          end: undefined,
        },
      };
      const resp = await axios.post(
        `${config.server.serverUrl}/api/v1/metrics-views/${metricsDef.id}/totals`,
        request
      );
      const totals = resp.data as MetricsViewTotalsResponse;
      assertBigNumber(totals.data, {
        measure_0: [300, 400],
      });
    });
  });
});
