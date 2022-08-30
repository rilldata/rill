import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExplorerActions";
import type { LeaderboardValues } from "$lib/redux-store/explore/explore-slice";
import axios from "axios";
import { MetricsExplorerTestData } from "../data/MetricsExplorer.data";
import { useBasicMetricsDefinition } from "../utils/metrics-definition-helpers";
import { normaliseLeaderboardOrder } from "../utils/normaliseLeaderboardOrder";
import { assertBigNumber } from "../utils/time-series-helpers";
import { useInlineTestServer } from "../utils/useInlineTestServer";

describe("Metrics Explore", () => {
  const { config, inlineServer } = useInlineTestServer(8083, 8084);
  let metricsDef: MetricsDefinitionEntity;
  let measures: Array<MeasureDefinitionEntity>;
  useBasicMetricsDefinition(inlineServer, (selMetricsDef, selMeasures) => {
    metricsDef = selMetricsDef;
    measures = selMeasures;
  });

  describe("Metrics leaderboard", () => {
    for (const MetricsExplorerTest of MetricsExplorerTestData) {
      it(`Should return leaderboard for ${MetricsExplorerTest.title}`, async () => {
        // select measures based on index passed or default to all measures
        const requestMeasures = MetricsExplorerTest.measures
          ? MetricsExplorerTest.measures.map((index) => measures[index])
          : measures;
        const resp = await axios.post(
          `${config.server.serverUrl}/api/metrics/${metricsDef.id}/leaderboards`,
          {
            measureId: requestMeasures[0].id,
            filters: MetricsExplorerTest.filters ?? {},
            ...(MetricsExplorerTest.timeRange
              ? { timeRange: MetricsExplorerTest.timeRange }
              : {}),
          }
        );
        const leaderboards = resp.data
          .split("\n")
          .filter((json) => !!json)
          .map(JSON.parse) as Array<LeaderboardValues>;
        expect(normaliseLeaderboardOrder(leaderboards)).toStrictEqual(
          MetricsExplorerTest.leaderboards
        );
      });
    }
  });

  describe("Metrics explore big number", () => {
    for (const MetricsExplorerTest of MetricsExplorerTestData) {
      it(`Should return big number for ${MetricsExplorerTest.title}`, async () => {
        // select measures based on index passed or default to all measures
        const requestMeasures = MetricsExplorerTest.measures
          ? MetricsExplorerTest.measures.map((index) => measures[index])
          : measures;
        const resp = await axios.post(
          `${config.server.serverUrl}/api/metrics/${metricsDef.id}/big-number`,
          {
            measures: requestMeasures,
            filters: MetricsExplorerTest.filters ?? {},
            ...(MetricsExplorerTest.timeRange
              ? { timeRange: MetricsExplorerTest.timeRange }
              : {}),
          }
        );
        const bigNumbers = resp.data as BigNumberResponse;
        assertBigNumber(bigNumbers, MetricsExplorerTest.bigNumber);
      });
    }
  });
});
