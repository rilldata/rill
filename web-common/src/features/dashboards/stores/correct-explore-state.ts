import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
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
 * Validates various fields in explore state.
 * Removes any invalid state. Cascading merge should fill in remaining state.
 *
 * Currently, it only acts on only a small section of the state.
 *
 * TODO: move all validations from convertUrlParamsToPreset and AdvancedMeasureCorrector here
 */
export function correctExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const correctedExploreState = { ...exploreState };
  const errors: Error[] = [];

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

  const errorsFromExploreView = correctExploreViewState(
    measures,
    dimensions,
    exploreSpec,
    correctedExploreState,
  );
  errors.push(...errorsFromExploreView);

  return { correctedExploreState, errors };
}

/**
 * Looks at any invalid fields in explore view and deletes it if it is completely invalid.
 */
function correctExploreViewState(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  exploreSpec: V1ExploreSpec,
  correctedExploreViewState: Partial<MetricsExplorerEntity>,
) {
  const errors: Error[] = [];

  if (correctedExploreViewState.visibleDimensions) {
    const selectedDimensions =
      correctedExploreViewState.visibleDimensions.filter((d) =>
        dimensions.has(d),
      );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      correctedExploreViewState.visibleDimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }

    if (selectedDimensions.length > 0) {
      correctedExploreViewState.allDimensionsVisible =
        selectedDimensions.length === exploreSpec.dimensions?.length;
      correctedExploreViewState.visibleDimensions = [...selectedDimensions];
    } else {
      delete correctedExploreViewState.allDimensionsVisible;
      delete correctedExploreViewState.visibleDimensions;
    }
  }

  // TODO: more validation once we need the full suite of validation
  return [
    ...errors,
    ...correctMeasureRelatedExploreViewState(
      measures,
      exploreSpec,
      correctedExploreViewState,
    ),
  ];
}

/**
 * Filters out invalid visible measures.
 * If all measures are invalid then it deletes the key for visible measures and any other settings based on visible measures.
 */
function correctMeasureRelatedExploreViewState(
  measures: Map<string, MetricsViewSpecMeasure>,
  exploreSpec: V1ExploreSpec,
  correctedExploreViewState: Partial<MetricsExplorerEntity>,
) {
  const errors: Error[] = [];

  let visibleMeasures = new Set();

  if (correctedExploreViewState.visibleMeasures) {
    const selectedMeasures = correctedExploreViewState.visibleMeasures.filter(
      (m) => measures.has(m),
    );
    const missingMeasures = getMissingValues(
      selectedMeasures,
      correctedExploreViewState.visibleMeasures,
    );
    if (missingMeasures.length) {
      errors.push(getMultiFieldError("measure", missingMeasures));
    }

    if (selectedMeasures.length > 0) {
      correctedExploreViewState.allMeasuresVisible =
        selectedMeasures.length === exploreSpec.measures?.length;
      correctedExploreViewState.visibleMeasures = [...selectedMeasures];
      visibleMeasures = new Set(selectedMeasures);
    } else {
      delete correctedExploreViewState.allMeasuresVisible;
      delete correctedExploreViewState.visibleMeasures;
    }
  }

  // No measure selected here was valid. So other measure setting based on these need to be unset as well.
  if (visibleMeasures.size === 0) {
    delete correctedExploreViewState.leaderboardSortByMeasureName;
    delete correctedExploreViewState.leaderboardMeasureNames;
    return errors;
  }

  if (
    correctedExploreViewState.leaderboardSortByMeasureName &&
    !visibleMeasures.has(correctedExploreViewState.leaderboardSortByMeasureName)
  ) {
    const measureIsPresentInMetricsView = measures.has(
      correctedExploreViewState.leaderboardSortByMeasureName,
    );
    errors.push(
      getSingleFieldError(
        "sort by measure",
        correctedExploreViewState.leaderboardSortByMeasureName,
        measureIsPresentInMetricsView ? "It is hidden." : "",
      ),
    );

    // Set the 1st valid sort measure if the selected measure is not valid
    correctedExploreViewState.leaderboardSortByMeasureName =
      correctedExploreViewState.visibleMeasures![0];
  }

  if (correctedExploreViewState.leaderboardMeasureNames?.length) {
    const selectedLeaderboardMeasures =
      correctedExploreViewState.leaderboardMeasureNames.filter((m) =>
        visibleMeasures.has(m),
      );
    const missingLeaderboardMeasures = getMissingValues(
      selectedLeaderboardMeasures,
      correctedExploreViewState.leaderboardMeasureNames,
    );
    if (missingLeaderboardMeasures.length) {
      errors.push(
        getMultiFieldError("leaderboard measure", missingLeaderboardMeasures),
      );
    }

    if (selectedLeaderboardMeasures.length > 0) {
      // If some measures are left after removing invalid measures then set those
      correctedExploreViewState.leaderboardMeasureNames = [
        ...selectedLeaderboardMeasures,
      ];
    } else {
      // Else set the 1st visible measure
      correctedExploreViewState.leaderboardMeasureNames = [
        correctedExploreViewState.visibleMeasures![0],
      ];
    }
  }

  return errors;
}
