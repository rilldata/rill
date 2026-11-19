import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import {
  isSimpleMeasure,
  measureSupportsTotalsQuery,
  visibleMeasures,
} from "./measures";
import type { DashboardDataSources } from "./types";

export const leaderboardSortByMeasureName = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardSortByMeasureName;
};

export const leaderboardMeasures = ({
  dashboard,
  ...rest
}: DashboardDataSources) => {
  const visibleMeasuresList = visibleMeasures({ dashboard, ...rest });
  const visibleMeasuresMap = getMapFromArray(
    visibleMeasuresList,
    (m) => m.name,
  );

  // Leaderboards group by each visible dimension in turn and also query totals,
  // so they can only query simple measures without required dimensions.
  // Filter and sort the leaderboard measure names based on the order in visibleMeasures
  const filteredMeasures = dashboard.leaderboardMeasureNames
    ?.map((name) => visibleMeasuresMap.get(name))
    ?.filter(
      (measure): measure is MetricsViewSpecMeasure =>
        !!measure && isLeaderboardMeasure(measure),
    )
    .sort((a, b) => {
      const aIndex = visibleMeasuresList.findIndex((m) => m.name === a.name);
      const bIndex = visibleMeasuresList.findIndex((m) => m.name === b.name);
      return aIndex - bIndex;
    });

  if (filteredMeasures?.length) return filteredMeasures;

  // Fall back to the sort measure, but never to one that leaderboards cannot query.
  const sortByMeasure = visibleMeasuresMap.get(
    dashboard.leaderboardSortByMeasureName,
  );
  if (sortByMeasure && isLeaderboardMeasure(sortByMeasure))
    return [sortByMeasure];

  return [];
};

const isLeaderboardMeasure = (measure: MetricsViewSpecMeasure) =>
  isSimpleMeasure(measure) && measureSupportsTotalsQuery(measure);

export const leaderboardMeasureNames = ({
  dashboard,
  ...rest
}: DashboardDataSources) => {
  const measures = leaderboardMeasures({ dashboard, ...rest });
  return measures.map((m) => m.name!);
};

export const leaderboardShowContextForAllMeasures = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardShowContextForAllMeasures ?? false;
};

export const leaderboardSelectors = {
  leaderboardSortByMeasureName,
  leaderboardMeasures,
  leaderboardMeasureNames,
  leaderboardShowContextForAllMeasures,
};
