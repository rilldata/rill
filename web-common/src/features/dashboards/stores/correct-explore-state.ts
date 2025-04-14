import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getMultiFieldError } from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Validates various fields in explore state. Correct any invalid state.
 * Currently, it acts on only a small section of the state.
 *
 * TODO: add extensive validation and move it to DashboardStateSync
 */
export function correctExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const measures = getMapFromArray(
    metricsViewSpec.measures?.filter((m) =>
      exploreSpec.measures?.includes(m.name!),
    ) ?? [],
    (m) => m.name!,
  );
  const dimensions = getMapFromArray(
    metricsViewSpec.dimensions?.filter((d) =>
      exploreSpec.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );

  correctExploreViewState(measures, dimensions, exploreSpec, exploreState);
}

function correctExploreViewState(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<MetricsExplorerEntity>,
) {
  if (exploreState.visibleMeasures) {
    const selectedMeasures = exploreState.visibleMeasures.filter((m) =>
      measures.has(m),
    );
    const missingMeasures = getMissingValues(
      selectedMeasures,
      exploreState.visibleMeasures,
    );
    if (missingMeasures.length) {
      // TODO: errors when we have validation in DashboardStateSync
    }

    exploreState.allMeasuresVisible =
      selectedMeasures.length === exploreSpec.measures?.length;
    exploreState.visibleMeasures = [...selectedMeasures];
  }

  if (exploreState.visibleDimensions) {
    const selectedDimensions = exploreState.visibleDimensions.filter((d) =>
      dimensions.has(d),
    );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      exploreState.visibleDimensions,
    );
    if (missingDimensions.length) {
      // TODO: errors when we have validation in DashboardStateSync
    }

    exploreState.allDimensionsVisible =
      selectedDimensions.length === exploreSpec.dimensions?.length;
    exploreState.visibleDimensions = [...selectedDimensions];
  }

  if (
    exploreState.leaderboardSortByMeasureName &&
    !measures.has(exploreState.leaderboardSortByMeasureName)
  ) {
    exploreState.leaderboardSortByMeasureName =
      exploreState.visibleMeasures?.[0];
  }

  // TODO: more validation once we need the full suite of validation
}
