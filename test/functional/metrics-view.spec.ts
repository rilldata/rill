import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type {
  MetricsViewTopListRequest,
  MetricsViewTotalsRequest,
  MetricsViewTotalsResponse,
} from "$common/rill-developer-service/MetricsViewActions";
import type { LeaderboardValues } from "$lib/application-state-stores/explorer-stores";
import axios from "axios";
import {
  DefaultCityLeaderboardReversed,
  MetricsExplorerTestData,
} from "../data/MetricsExplorer.data";
import { useBasicMetricsDefinition } from "../utils/metrics-definition-helpers";
import { normaliseLeaderboardOrder } from "../utils/normaliseLeaderboardOrder";
import { assertBigNumber } from "../utils/time-series-helpers";
import { useInlineTestServer } from "../utils/useInlineTestServer";

describe("Metrics View", () => {
  const { config, inlineServer } = useInlineTestServer(8083);
  let metricsDef: MetricsDefinitionEntity;
  let measures: Array<MeasureDefinitionEntity>;
  let dimensions: Array<DimensionDefinitionEntity>;
  useBasicMetricsDefinition(
    inlineServer,
    (selMetricsDef, selMeasures, selDimensions) => {
      metricsDef = selMetricsDef;
      measures = selMeasures;
      dimensions = selDimensions;
    }
  );

  describe("Top List", () => {
    for (const MetricsExplorerTest of MetricsExplorerTestData) {
      it(`Should return top list for ${MetricsExplorerTest.title}`, async () => {
        // select measures based on index passed or default to all measures
        const requestMeasures = MetricsExplorerTest.measures
          ? MetricsExplorerTest.measures.map((index) => measures[index])
          : measures;

        const leaderboards = await Promise.all(
          dimensions.map(async (dimension) => {
            const request: MetricsViewTopListRequest = {
              measures: [requestMeasures[0].id],
              filter: MetricsExplorerTest.filters,
              time: {
                start: MetricsExplorerTest.timeRange?.start,
                end: MetricsExplorerTest.timeRange?.end,
              },
              limit: 15,
              offset: 0,
              sort: [{ name: requestMeasures[0].sqlName, direction: "desc" }],
            };
            const resp = await axios.post(
              `${config.server.serverUrl}/api/v1/metrics-views/${metricsDef.id}/toplist/${dimension.id}`,
              request,
              { responseType: "json" }
            );
            return {
              values: resp.data.data,
              dimensionName: dimension.dimensionColumn,
            } as LeaderboardValues;
          })
        );

        expect(
          normaliseLeaderboardOrder(leaderboards, requestMeasures[0].sqlName)
        ).toStrictEqual(MetricsExplorerTest.leaderboards);
      });
    }
  });

  it("Top list with multiple measures and sorting", async () => {
    const dimension = dimensions[2];
    const request: MetricsViewTopListRequest = {
      measures: measures.map((measure) => measure.id),
      time: { start: undefined, end: undefined },
      limit: 15,
      offset: 0,
      sort: [{ name: measures[0].sqlName, direction: "asc" }],
    };
    const resp = await axios.post(
      `${config.server.serverUrl}/api/v1/metrics-views/${metricsDef.id}/toplist/${dimension.id}`,
      request,
      { responseType: "json" }
    );

    // 2nd measure is present
    expect(resp.data.data[0][measures[1].sqlName]).toBeGreaterThan(0);
    // check 1st measure in reverse order
    expect(
      normaliseLeaderboardOrder(
        [
          {
            values: resp.data.data,
            dimensionName: dimension.dimensionColumn,
          } as LeaderboardValues,
        ],
        measures[0].sqlName
      )
    ).toStrictEqual([DefaultCityLeaderboardReversed]);
  });

  describe("Metrics view totals", () => {
    for (const MetricsExplorerTest of MetricsExplorerTestData) {
      it(`Should return totals for ${MetricsExplorerTest.title}`, async () => {
        const request: MetricsViewTotalsRequest = {
          // select measures based on index passed or default to all measures
          measures: MetricsExplorerTest.measures
            ? MetricsExplorerTest.measures.map((index) => measures[index].id)
            : measures.map((measure) => measure.id),
          filter: MetricsExplorerTest.filters,
          time: {
            start: MetricsExplorerTest.timeRange?.start,
            end: MetricsExplorerTest.timeRange?.end,
          },
        };
        const resp = await axios.post(
          `${config.server.serverUrl}/api/v1/metrics-views/${metricsDef.id}/totals`,
          request
        );
        const totals = resp.data as MetricsViewTotalsResponse;
        assertBigNumber(totals.data, MetricsExplorerTest.bigNumber);
      });
    }
  });
});
