import { useInlineTestServer } from "../utils/useInlineTestServer";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { useBasicMetricsDefinition } from "../utils/metrics-definition-helpers";
import { MetricsExploreTestData } from "../data/MetricsExplore.data";
import type { LeaderboardValues } from "$lib/redux-store/explore/explore-slice";
import axios from "axios";
import { normaliseLeaderboardOrder } from "../utils/normaliseLeaderboardOrder";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExploreActions";
import { assertBigNumber } from "../utils/time-series-helpers";

describe("Metrics Explore", () => {
  const { config, inlineServer } = useInlineTestServer(8083);
  let metricsDef: MetricsDefinitionEntity;
  let measures: Array<MeasureDefinitionEntity>;
  useBasicMetricsDefinition(inlineServer, (selMetricsDef, selMeasures) => {
    metricsDef = selMetricsDef;
    measures = selMeasures;
  });

  describe("Metrics leaderboard", () => {
    for (const MetricsExploreTest of MetricsExploreTestData) {
      it(`Should return leaderboard for ${MetricsExploreTest.title}`, async () => {
        // select measures based on index passed or default to all measures
        const requestMeasures = MetricsExploreTest.measures
          ? MetricsExploreTest.measures.map((index) => measures[index])
          : measures;
        const resp = await axios.post(
          `${config.server.serverUrl}/api/metrics/${metricsDef.id}/leaderboards`,
          {
            measureId: requestMeasures[0].id,
            filters: MetricsExploreTest.filters ?? {},
            ...(MetricsExploreTest.timeRange
              ? { timeRange: MetricsExploreTest.timeRange }
              : {}),
          }
        );
        const leaderboards = resp.data
          .split("\n")
          .filter((json) => !!json)
          .map(JSON.parse) as Array<LeaderboardValues>;
        console.log(normaliseLeaderboardOrder(leaderboards));
        expect(normaliseLeaderboardOrder(leaderboards)).toStrictEqual(
          MetricsExploreTest.leaderboards
        );
      });
    }
  });

  describe("Metrics explore big number", () => {
    for (const MetricsExploreTest of MetricsExploreTestData) {
      it(`Should return big number for ${MetricsExploreTest.title}`, async () => {
        // select measures based on index passed or default to all measures
        const requestMeasures = MetricsExploreTest.measures
          ? MetricsExploreTest.measures.map((index) => measures[index])
          : measures;
        const resp = await axios.post(
          `${config.server.serverUrl}/api/metrics/${metricsDef.id}/big-number`,
          {
            measures: requestMeasures,
            filters: MetricsExploreTest.filters ?? {},
            ...(MetricsExploreTest.timeRange
              ? { timeRange: MetricsExploreTest.timeRange }
              : {}),
          }
        );
        const bigNumbers = resp.data as BigNumberResponse;
        console.log(bigNumbers);
        assertBigNumber(bigNumbers, MetricsExploreTest.bigNumber);
      });
    }
  });
});
