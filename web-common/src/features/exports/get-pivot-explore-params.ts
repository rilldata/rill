import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import {
  measureLabel,
  visibleMeasures,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
import type { DashboardDataSources } from "@rilldata/web-common/features/dashboards/state-managers/selectors/types.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import type { TimeGrain } from "@rilldata/web-common/lib/time/types.ts";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { getDimensionDisplayName } from "../dashboards/state-managers/selectors/dimensions";

export function getPivotExploreParams(
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeControlState: TimeControlState,
) {
  const newExploreState = { ...exploreState };

  switch (exploreState.activePage) {
    case DashboardState_ActivePage.PIVOT:
      break;

    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
      // There is no option to create a report from the default view so we can keep the pivot state empty for now.
      break;

    case DashboardState_ActivePage.DIMENSION_TABLE:
      newExploreState.pivot = getPivotStateForDimensionTable(
        exploreState,
        metricsViewSpec,
        exploreSpec,
      );
      break;

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      newExploreState.pivot = getPivotStateForTDD(
        exploreState,
        metricsViewSpec,
        exploreSpec,
      );
      break;
  }

  newExploreState.activePage = DashboardState_ActivePage.PIVOT;

  const urlSearchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    newExploreState,
    timeControlState,
  );

  return urlSearchParams;
}

function getPivotStateForDimensionTable(
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
) {
  if (!exploreState.selectedDimensionName) return exploreState.pivot;

  const dashboardDataSources = {
    dashboard: exploreState,
    validMetricsView: metricsViewSpec,
    validExplore: exploreSpec,
  } as DashboardDataSources;

  const expandedDimensionTitle = getDimensionDisplayName(dashboardDataSources)(
    exploreState.selectedDimensionName,
  );

  const rows = [
    {
      id: exploreState.selectedDimensionName,
      title: expandedDimensionTitle,
      type: PivotChipType.Dimension,
    },
  ];

  const curVisibleMeasures = visibleMeasures(dashboardDataSources);
  const columns = curVisibleMeasures
    .filter((m) => m.name !== undefined)
    .map((m) => {
      return {
        id: m.name as string,
        title: m.displayName || (m.name as string),
        type: PivotChipType.Measure,
      };
    });

  return {
    ...exploreState.pivot,
    rows,
    columns,
  };
}

function getPivotStateForTDD(
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
) {
  const dashboardGrain = exploreState.selectedTimeRange?.interval;
  if (!dashboardGrain || !exploreState.tdd.expandedMeasureName)
    return exploreState.pivot;

  const dashboardDataSources = {
    dashboard: exploreState,
    validMetricsView: metricsViewSpec,
    validExplore: exploreSpec,
  } as DashboardDataSources;

  const rows: PivotChipData[] = [];
  if (exploreState.selectedComparisonDimension) {
    const compareDimensionTitle = getDimensionDisplayName(dashboardDataSources)(
      exploreState.selectedComparisonDimension,
    );
    rows.push({
      id: exploreState.selectedComparisonDimension,
      title: compareDimensionTitle,
      type: PivotChipType.Dimension,
    });
  }

  const timeGrain: TimeGrain = TIME_GRAIN[dashboardGrain];
  const measureTitle = measureLabel(dashboardDataSources)(
    exploreState.tdd.expandedMeasureName,
  );
  const columns: PivotChipData[] = [
    {
      id: dashboardGrain,
      title: timeGrain.label,
      type: PivotChipType.Time,
    },
    {
      id: exploreState.tdd.expandedMeasureName,
      title: measureTitle,
      type: PivotChipType.Measure,
    },
  ];

  return {
    ...exploreState.pivot,
    rows,
    columns,
  };
}
