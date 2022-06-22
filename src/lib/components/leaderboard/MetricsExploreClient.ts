import { HttpStreamClient } from "$lib/http-client/HttpStreamClient";
import type { ActiveValues } from "$lib/redux-store/metrics-leaderboard-slice";
import { config } from "$lib/application-state-stores/application-store";

export class MetricsExploreClient {
  public static async getLeaderboardValues(
    metricsDefId: string,
    measureId: string,
    filters: ActiveValues
  ) {
    return HttpStreamClient.instance.request(
      `/metrics/${metricsDefId}/leaderboards`,
      "POST",
      { measureId, filters }
    );
  }

  public static async getBigNumber(
    metricsDefId: string,
    measureId: string,
    filters: ActiveValues
  ): Promise<number> {
    return (
      await (
        await fetch(
          `${config.server.serverUrl}/api/metrics/${metricsDefId}/bigNumber`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ measureId, filters }),
          }
        )
      ).json()
    ).data;
  }
}
