import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type {
  MetricViewTopListRequest,
  MetricViewTotalsRequest,
  MetricViewTotalsResponse,
} from "$common/rill-developer-service/MetricViewActions";
import type { LeaderboardValues } from "$lib/application-state-stores/explorer-stores";
import axios from "axios";
import { MetricsExplorerTestData } from "../data/MetricsExplorer.data";
import { useBasicMetricsDefinition } from "../utils/metrics-definition-helpers";
import { normaliseLeaderboardOrder } from "../utils/normaliseLeaderboardOrder";
import { assertBigNumber } from "../utils/time-series-helpers";
import { useInlineTestServer } from "../utils/useInlineTestServer";

describe("Metric View", () => {
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
            const request: MetricViewTopListRequest = {
              measures: [requestMeasures[0].id],
              filter: MetricsExplorerTest.filters,
              time: {
                start: MetricsExplorerTest.timeRange?.start,
                end: MetricsExplorerTest.timeRange?.end,
              },
              limit: 0,
              offset: 0,
              sort: [],
            };
            const resp = await axios.post(
              `${config.server.serverUrl}/api/v1/metric-views/${metricsDef.id}/toplist/${dimension.id}`,
              request,
              { responseType: "json" }
            );
            return {
              values: resp.data.data,
              dimensionName: dimension.dimensionColumn,
            } as LeaderboardValues;
          })
        );

        expect(normaliseLeaderboardOrder(leaderboards)).toStrictEqual(
          MetricsExplorerTest.leaderboards
        );
      });
    }
  });

  describe("Metric view totals", () => {
    for (const MetricsExplorerTest of MetricsExplorerTestData) {
      it(`Should return totals for ${MetricsExplorerTest.title}`, async () => {
        const request: MetricViewTotalsRequest = {
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
          `${config.server.serverUrl}/api/v1/metric-views/${metricsDef.id}/totals`,
          request
        );
        const totals = resp.data as MetricViewTotalsResponse;
        assertBigNumber(totals.data, MetricsExplorerTest.bigNumber);
      });
    }
  });
});
