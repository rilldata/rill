import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
import {
  type MetricsViewSpecMeasure,
  MetricsViewSpecMeasureType,
} from "@rilldata/web-common/runtime-client";
import { visibleMeasures } from "./measures";
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

  // Filter and sort the leaderboard measure names based on the order in visibleMeasures
  const filteredMeasures = dashboard.leaderboardMeasureNames
    ?.map((name) => visibleMeasuresMap.get(name))
    ?.filter((measure) => {
      if (!measure) return false;
      return (
        measure.type !==
          MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
        !measure.window
      );
    })
    .sort((a, b) => {
      const aIndex = visibleMeasuresList.findIndex((m) => m.name === a?.name);
      const bIndex = visibleMeasuresList.findIndex((m) => m.name === b?.name);
      return aIndex - bIndex;
    }) as MetricsViewSpecMeasure[];

  if (filteredMeasures?.length) return filteredMeasures;

  if (visibleMeasuresMap.has(dashboard.leaderboardSortByMeasureName))
    return [
      visibleMeasuresMap.get(
        dashboard.leaderboardSortByMeasureName,
      ) as MetricsViewSpecMeasure,
    ];

  return [];
};

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
