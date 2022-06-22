import { MetricsExploreClient } from "$lib/components/leaderboard/MetricsExploreClient";
import { store } from "$lib/redux-store/store-root";
import {
  ActiveValues,
  MetricsLeaderboardEntity,
  setBigNumber,
  setReferenceValue,
} from "$lib/redux-store/metrics-leaderboard-slice";

// prepare the activeFilters to be sent to the server
function prune(actives) {
  const filters: ActiveValues = {};
  for (const activeColumn in actives) {
    if (!actives[activeColumn].length) continue;
    filters[activeColumn] = actives[activeColumn];
  }
  return filters;
}

export function updateDisplay(
  metricsDefId: string,
  metricsLeaderboard: MetricsLeaderboardEntity,
  anythingSelected: boolean
) {
  const filters = prune(metricsLeaderboard.activeValues);
  MetricsExploreClient.getLeaderboardValues(
    metricsDefId,
    metricsLeaderboard.measureId,
    filters
  );
  MetricsExploreClient.getBigNumber(
    metricsDefId,
    metricsLeaderboard.measureId,
    filters
  ).then((bigNumber) => {
    store.dispatch(setBigNumber(metricsDefId, bigNumber));

    if (anythingSelected) {
      store.dispatch(setReferenceValue(metricsDefId, bigNumber));
    }
  });
}
